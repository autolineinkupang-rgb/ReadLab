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
	URL         string `json:"url" binding:"required"`
	MaxChapters int    `json:"max_chapters"`
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

	slug := generateSlug(result.Title)

	novel := model.Novel{
		Title:       result.Title,
		Slug:        slug,
		Author:      result.Author,
		AuthorSlug:  generateSlug(result.Author),
		Status:      "ongoing",
		Description: "Imported via lncrawl.",
		CoverURL:    result.CoverURL,
		Chapters:    result.Total,
	}

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&novel).Error; err != nil {
			return err
		}
		chapters := make([]model.Chapter, 0, len(result.Chapters))
		for _, ch := range result.Chapters {
			chapters = append(chapters, model.Chapter{
				NovelID: novel.ID,
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
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "import failed: " + err.Error()})
		return
	}

	h.DB.Preload("Genres").First(&novel, novel.ID)
	c.JSON(http.StatusCreated, gin.H{"data": novel})
}
