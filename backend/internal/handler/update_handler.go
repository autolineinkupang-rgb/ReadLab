package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/model"
)

type UpdateHandler struct {
	DB *gorm.DB
}

func NewUpdateHandler(db *gorm.DB) *UpdateHandler {
	return &UpdateHandler{DB: db}
}

func (h *UpdateHandler) Recent(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	if limit < 1 || limit > 100 {
		limit = 30
	}

	var updates []model.Chapter
	if err := h.DB.Preload("Novel").
		Order("created_at DESC").
		Limit(limit).
		Find(&updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": updates})
}
