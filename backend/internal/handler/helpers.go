package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/ticket"
)

func mustUserID(c *gin.Context) (uint, bool) {
	raw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return 0, false
	}
	uid, ok := raw.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return 0, false
	}
	return uid, true
}

func respondErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	case errors.Is(err, ticket.ErrInsufficientTickets):
		c.JSON(http.StatusPaymentRequired, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
