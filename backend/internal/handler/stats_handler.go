package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"readlab/backend/internal/model"
)

type StatsHandler struct {
	DB *gorm.DB
}

func NewStatsHandler(db *gorm.DB) *StatsHandler {
	return &StatsHandler{DB: db}
}

func (h *StatsHandler) Get(c *gin.Context) {
	var novelCount, chapterCount, userCount, voteCount, requestCount int64
	var totalViews uint64

	h.DB.Model(&model.Novel{}).Count(&novelCount)
	h.DB.Model(&model.Chapter{}).Count(&chapterCount)
	h.DB.Model(&model.User{}).Count(&userCount)
	h.DB.Model(&model.Vote{}).Count(&voteCount)
	h.DB.Model(&model.Request{}).Count(&requestCount)

	h.DB.Model(&model.Novel{}).Select("COALESCE(SUM(views), 0)").Scan(&totalViews)

	c.JSON(http.StatusOK, gin.H{
		"total_novels":   novelCount,
		"total_chapters": chapterCount,
		"total_users":    userCount,
		"total_views":    totalViews,
		"total_votes":    voteCount,
		"total_requests": requestCount,
	})
}
