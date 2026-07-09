package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/lncrawl"
	"wtr-lab-clone/backend/internal/model"
)

type LncrawlHandler struct {
	DB *gorm.DB
}

func NewLncrawlHandler(db *gorm.DB) *LncrawlHandler {
	return &LncrawlHandler{DB: db}
}

type LncrawlRequest struct {
	URL          string `json:"url" binding:"required"`
	MaxChapters  int    `json:"max_chapters"`
	ChapterRange string `json:"chapter_range"`
}

func (h *LncrawlHandler) Crawl(c *gin.Context) {
	var req LncrawlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := lncrawl.RunCrawl(req.URL, req.MaxChapters)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "crawl failed: " + err.Error()})
		return
	}

	chapterFilter := parseChapterRange(req.ChapterRange)
	if len(chapterFilter) > 0 {
		var filtered []lncrawl.ChapterContent
		for _, ch := range result.Chapters {
			if chapterFilter[ch.Number] {
				filtered = append(filtered, ch)
			}
		}
		result.Chapters = filtered
	}

	var existingNovel model.Novel
	exists := h.DB.Where("source_url = ?", req.URL).First(&existingNovel).Error == nil
	if !exists {
		exists = h.DB.Where("LOWER(title) = LOWER(?)", result.Title).First(&existingNovel).Error == nil
	}

	var novelID uint
	var isNew bool

	if exists {
		novelID = existingNovel.ID
		isNew = false
	} else {
		slug := generateSlug(result.Title)
		dbNovel := model.Novel{
			Title:       result.Title,
			Slug:        slug,
			Author:      result.Author,
			AuthorSlug:  generateSlug(result.Author),
			Status:      "ongoing",
			Description: "Imported via lncrawl.",
			CoverURL:    result.CoverURL,
			SourceURL:   req.URL,
			Chapters:    0,
		}
		if err := h.DB.Create(&dbNovel).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create novel: " + err.Error()})
			return
		}
		novelID = dbNovel.ID
		isNew = true
	}

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		if isNew {
			chapters := make([]model.Chapter, 0, len(result.Chapters))
			for _, ch := range result.Chapters {
				chapters = append(chapters, model.Chapter{
					NovelID: novelID,
					Number:  ch.Number,
					Title:   ch.Title,
					Content: ch.Content,
				})
			}
			if len(chapters) > 0 {
				if err := tx.Create(&chapters).Error; err != nil {
					return err
				}
			}
		} else {
			updates := map[string]interface{}{}
			if existingNovel.CoverURL == "" && result.CoverURL != "" {
				updates["cover_url"] = result.CoverURL
			}
			if len(updates) > 0 {
				tx.Model(&model.Novel{}).Where("id = ?", novelID).Updates(updates)
			}
			for _, ch := range result.Chapters {
				var count int64
				tx.Model(&model.Chapter{}).Where("novel_id = ? AND number = ?", novelID, ch.Number).Count(&count)
				if count == 0 {
					if err := tx.Create(&model.Chapter{
						NovelID: novelID,
						Number:  ch.Number,
						Title:   ch.Title,
						Content: ch.Content,
					}).Error; err != nil {
						return err
					}
				}
			}
		}

		var totalChapters int64
		tx.Model(&model.Chapter{}).Where("novel_id = ?", novelID).Count(&totalChapters)
		tx.Model(&model.Novel{}).Where("id = ?", novelID).UpdateColumn("chapters", totalChapters)

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "import failed: " + err.Error()})
		return
	}

	var novel model.Novel
	h.DB.Preload("Genres").Preload("Tags").First(&novel, novelID)
	statusCode := http.StatusCreated
	if !isNew {
		statusCode = http.StatusOK
	}
	c.JSON(statusCode, gin.H{"data": novel})
}
