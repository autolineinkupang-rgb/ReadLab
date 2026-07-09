package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/model"
)

type AuthorHandler struct {
	DB *gorm.DB
}

func NewAuthorHandler(db *gorm.DB) *AuthorHandler {
	return &AuthorHandler{DB: db}
}

func (h *AuthorHandler) Novels(c *gin.Context) {
	authorName := c.Param("name")

	var novels []model.Novel
	if err := h.DB.Preload("Genres").Preload("Tags").
		Where("author = ?", authorName).
		Order("created_at DESC").
		Find(&novels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  novels,
		"total": len(novels),
	})
}
