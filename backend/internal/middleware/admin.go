package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AdminRequired(_ *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}
		c.Next()
	}
}
