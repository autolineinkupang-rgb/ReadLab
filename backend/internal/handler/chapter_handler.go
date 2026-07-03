package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/model"
)

type ChapterHandler struct {
	DB *gorm.DB
}

func NewChapterHandler(db *gorm.DB) *ChapterHandler {
	return &ChapterHandler{DB: db}
}

func (h *ChapterHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var chapter model.Chapter
	if err := h.DB.Preload("Novel").First(&chapter, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "chapter not found"})
		return
	}

	c.JSON(http.StatusOK, chapter)
}
