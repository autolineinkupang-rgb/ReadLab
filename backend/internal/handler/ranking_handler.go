package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/model"
)

type RankingHandler struct {
	DB *gorm.DB
}

func NewRankingHandler(db *gorm.DB) *RankingHandler {
	return &RankingHandler{DB: db}
}

func (h *RankingHandler) Get(c *gin.Context) {
	period := c.Param("period")

	var novels []model.Novel
	query := h.DB.Preload("Genres").Order("views DESC").Limit(50)

	if period != "all_time" {
		query = query.Where("created_at > NOW() - INTERVAL '1 " + period + "'")
	}

	if err := query.Find(&novels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": novels})
}
