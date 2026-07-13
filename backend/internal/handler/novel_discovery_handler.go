package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"readlab/backend/internal/model"
)

func (h *NovelHandler) Random(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	var novels []model.Novel
	var maxID uint
	h.DB.Model(&model.Novel{}).Select("MAX(id)").Scan(&maxID)
	if maxID == 0 {
		c.JSON(http.StatusOK, gin.H{"data": []model.Novel{}})
		return
	}

	if err := h.DB.Preload("Genres").Preload("Tags").
		Where("id >= FLOOR(RANDOM() * ?) + 1", maxID).
		Limit(limit).
		Find(&novels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": novels})
}

func (h *NovelHandler) Trending(c *gin.Context) {
	var novels []model.Novel
	if err := h.DB.Preload("Genres").Preload("Tags").
		Order("views DESC").
		Limit(20).
		Find(&novels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": novels})
}

func (h *NovelHandler) Recommendations(c *gin.Context) {
	var novels []model.Novel
	if err := h.DB.Preload("Genres").Preload("Tags").
		Order("rating DESC, views DESC").
		Limit(12).
		Find(&novels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": novels})
}