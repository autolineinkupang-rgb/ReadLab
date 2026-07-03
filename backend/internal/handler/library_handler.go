package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/model"
)

type LibraryHandler struct {
	DB *gorm.DB
}

func NewLibraryHandler(db *gorm.DB) *LibraryHandler {
	return &LibraryHandler{DB: db}
}

func (h *LibraryHandler) Get(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var follows []model.NovelFollow
	if err := h.DB.Where("user_id = ?", userID).
		Preload("Novel").
		Find(&follows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var history []model.ReadingHistory
	if err := h.DB.Where("user_id = ?", userID).
		Preload("Novel").
		Preload("Chapter").
		Order("updated_at DESC").
		Limit(50).
		Find(&history).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"follows": follows,
		"history": history,
	})
}
