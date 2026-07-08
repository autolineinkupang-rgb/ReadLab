package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/model"
)

type LeaderboardHandler struct {
	DB *gorm.DB
}

func NewLeaderboardHandler(db *gorm.DB) *LeaderboardHandler {
	return &LeaderboardHandler{DB: db}
}

func (h *LeaderboardHandler) Get(c *gin.Context) {
	sort := c.DefaultQuery("sort", "xp")

	var users []model.User
	query := h.DB

	switch sort {
	case "ticket_count":
		query = query.Where("tickets > 0").Order("tickets DESC")
	default:
		query = query.Where("xp > 0").Order("xp DESC")
	}

	if err := query.Limit(100).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": users})
}
