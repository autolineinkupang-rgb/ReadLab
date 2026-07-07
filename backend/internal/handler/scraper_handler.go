package handler

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/model"
	"wtr-lab-clone/backend/internal/scraper"
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
		Chapters:    len(novel.Chapters),
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
		if err := tx.Create(&dbNovel).Error; err != nil {
			return err
		}
		if len(matchedGenres) > 0 {
			if err := tx.Model(&dbNovel).Association("Genres").Append(matchedGenres); err != nil {
				return err
			}
		}

		type chapterJob struct {
			ch  scraper.ScrapedChapter
			err error
		}

		chapters := make([]model.Chapter, 0, len(novel.Chapters))
		var mu sync.Mutex

		if req.WithContent {
			var wg sync.WaitGroup
			ch := make(chan scraper.ScrapedChapter, len(novel.Chapters))

			for _, sc := range novel.Chapters {
				wg.Add(1)
				go func(sc scraper.ScrapedChapter) {
					defer wg.Done()
					chapter, err := h.Scraper.ScrapeChapter(sc.URL)
					if err == nil && chapter.Content != "" {
						mu.Lock()
						chapters = append(chapters, model.Chapter{
							NovelID: dbNovel.ID,
							Number:  chapter.Number,
							Title:   chapter.Title,
							Content: chapter.Content,
						})
						mu.Unlock()
					} else {
						mu.Lock()
						chapters = append(chapters, model.Chapter{
							NovelID: dbNovel.ID,
							Number:  sc.Number,
							Title:   sc.Title,
							Content: "",
						})
						mu.Unlock()
					}
				}(sc)
			}
			wg.Wait()
			close(ch)
		} else {
			for _, sc := range novel.Chapters {
				chapters = append(chapters, model.Chapter{
					NovelID: dbNovel.ID,
					Number:  sc.Number,
					Title:   sc.Title,
				})
			}
		}

		if len(chapters) > 0 {
			if err := tx.Create(&chapters).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "import failed: " + err.Error()})
		return
	}

	h.DB.Preload("Genres").First(&dbNovel, dbNovel.ID)
	c.JSON(http.StatusCreated, gin.H{"data": dbNovel})
}
