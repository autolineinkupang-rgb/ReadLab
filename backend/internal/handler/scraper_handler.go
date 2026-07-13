package handler

import (
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"readlab/backend/internal/model"
	"readlab/backend/internal/scraper"
)

type ScraperHandler struct {
	DB      *gorm.DB
	Scraper *scraper.Scraper
}

func NewScraperHandler(db *gorm.DB) *ScraperHandler {
	return &ScraperHandler{
		DB:      db,
		Scraper: scraper.New(),
	}
}

type ScrapeRequest struct {
	URL string `json:"url" binding:"required"`
}

func (h *ScraperHandler) Scrape(c *gin.Context) {
	var req ScrapeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	novel, err := h.Scraper.ScrapeNovel(req.URL)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "scrape failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": novel})
}

type ScrapeImportRequest struct {
	URL          string `json:"url" binding:"required"`
	WithContent  bool   `json:"with_content"`
	ChapterRange string `json:"chapter_range"`
}

func parseChapterRange(rangeStr string) map[int]bool {
	result := make(map[int]bool)
	if rangeStr == "" || rangeStr == "all" {
		return result
	}
	for _, part := range strings.Split(rangeStr, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			start, err1 := strconv.Atoi(strings.TrimSpace(bounds[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(bounds[1]))
			if err1 == nil && err2 == nil && start > 0 && end >= start {
				for i := start; i <= end; i++ {
					result[i] = true
				}
			}
		} else {
			num, err := strconv.Atoi(part)
			if err == nil && num > 0 {
				result[num] = true
			}
		}
	}
	return result
}

func (h *ScraperHandler) Import(c *gin.Context) {
	var req ScrapeImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	novel, err := h.Scraper.ScrapeNovel(req.URL)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "scrape failed: " + err.Error()})
		return
	}

	var existingNovel model.Novel
	exists := h.DB.Where("source_url = ?", req.URL).First(&existingNovel).Error == nil
	if !exists {
		exists = h.DB.Where("LOWER(title) = LOWER(?)", novel.Title).First(&existingNovel).Error == nil
	}

	var dbNovelID uint
	var isNew bool

	if exists {
		dbNovelID = existingNovel.ID
		isNew = false
	} else {
		slug := generateSlug(novel.Title)
		status := novel.Status
		if status == "" {
			status = "ongoing"
		}
		dbNovel := model.Novel{
			Title:       novel.Title,
			AltTitle:    novel.AltTitle,
			Slug:        slug,
			Author:      novel.Author,
			AuthorSlug:  generateSlug(novel.Author),
			Status:      status,
			Description: novel.Description,
			CoverURL:    novel.CoverURL,
			SourceURL:   req.URL,
			Chapters:    0,
		}
		if err := h.DB.Create(&dbNovel).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create novel: " + err.Error()})
			return
		}
		dbNovelID = dbNovel.ID
		isNew = true
	}

	chapterFilter := parseChapterRange(req.ChapterRange)
	if len(chapterFilter) > 0 {
		var filtered []scraper.ScrapedChapter
		for _, ch := range novel.Chapters {
			if chapterFilter[ch.Number] {
				filtered = append(filtered, ch)
			}
		}
		novel.Chapters = filtered
	}

	var matchedGenres []model.Genre
	for _, gName := range novel.Genres {
		slug := generateSlug(gName)
		var genre model.Genre
		err := h.DB.Where("LOWER(name) = ? OR slug = ?", gName, slug).First(&genre).Error
		if err != nil {
			genre = model.Genre{Slug: slug, Name: gName}
			if err := h.DB.Create(&genre).Error; err != nil {
				continue
			}
		}
		matchedGenres = append(matchedGenres, genre)
	}

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		if isNew {
			if len(matchedGenres) > 0 {
				var n model.Novel
				if err := tx.First(&n, dbNovelID).Error; err != nil {
					return err
				}
				if err := tx.Model(&n).Association("Genres").Append(matchedGenres); err != nil {
					return err
				}
			}
		} else {
			status := novel.Status
			if status == "" {
				status = "ongoing"
			}
			updates := map[string]interface{}{
				"cover_url":   novel.CoverURL,
				"description": novel.Description,
				"author":      novel.Author,
				"author_slug": generateSlug(novel.Author),
				"alt_title":   novel.AltTitle,
				"status":      status,
			}
			tx.Model(&model.Novel{}).Where("id = ?", dbNovelID).Updates(updates)
			if len(matchedGenres) > 0 {
				var n model.Novel
				if err := tx.First(&n, dbNovelID).Error; err != nil {
					return err
				}
				if err := tx.Model(&n).Association("Genres").Replace(matchedGenres); err != nil {
					return err
				}
			}
		}

		chapters := make([]model.Chapter, 0, len(novel.Chapters))
		var mu sync.Mutex

		if req.WithContent {
			var wg sync.WaitGroup

			for _, sc := range novel.Chapters {
				wg.Add(1)
				go func(sc scraper.ScrapedChapter) {
					defer wg.Done()
					chapter, err := h.Scraper.ScrapeChapter(sc.URL)
					mu.Lock()
					if err == nil && chapter.Content != "" {
						chapters = append(chapters, model.Chapter{
							NovelID: dbNovelID,
							Number:  chapter.Number,
							Title:   chapter.Title,
							Content: chapter.Content,
						})
					} else {
						chapters = append(chapters, model.Chapter{
							NovelID: dbNovelID,
							Number:  sc.Number,
							Title:   sc.Title,
							Content: "",
						})
					}
					mu.Unlock()
				}(sc)
			}
			wg.Wait()
		} else {
			for _, sc := range novel.Chapters {
				chapters = append(chapters, model.Chapter{
					NovelID: dbNovelID,
					Number:  sc.Number,
					Title:   sc.Title,
				})
			}
		}

		for _, ch := range chapters {
			var count int64
			tx.Model(&model.Chapter{}).Where("novel_id = ? AND number = ?", dbNovelID, ch.Number).Count(&count)
			if count == 0 {
				if err := tx.Create(&ch).Error; err != nil {
					return err
				}
			}
		}

		var totalChapters int64
		tx.Model(&model.Chapter{}).Where("novel_id = ?", dbNovelID).Count(&totalChapters)
		tx.Model(&model.Novel{}).Where("id = ?", dbNovelID).UpdateColumn("chapters", totalChapters)

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "import failed: " + err.Error()})
		return
	}

	var result model.Novel
	h.DB.Preload("Genres").Preload("Tags").First(&result, dbNovelID)
	statusCode := http.StatusCreated
	if !isNew {
		statusCode = http.StatusOK
	}
	c.JSON(statusCode, gin.H{"data": result})
}
