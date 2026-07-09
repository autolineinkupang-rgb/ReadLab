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

var validPeriods = map[string]string{
	"daily":   "day",
	"weekly":  "week",
	"monthly": "month",
}

func (h *RankingHandler) Get(c *gin.Context) {
	period := c.Param("period")

	var novels []model.Novel
	query := h.DB.Preload("Genres").Order("views DESC").Limit(50)

	if interval, ok := validPeriods[period]; ok {
		query = query.Where("created_at > NOW() - CAST(? AS INTERVAL)", "1 "+interval)
	} else if period != "all_time" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period: " + period})
		return
	}

	if err := query.Find(&novels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": novels})
}
