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
	sort := c.DefaultQuery("sort", "ticket_count")

	var users []model.User
	query := h.DB.Where("tickets > 0")

	if sort == "ticket_count" {
		query = query.Order("tickets DESC")
	}

	if err := query.Limit(100).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": users})
}
