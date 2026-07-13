package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"readlab/backend/internal/model"
)

type GenreHandler struct {
	DB *gorm.DB
}

func NewGenreHandler(db *gorm.DB) *GenreHandler {
	return &GenreHandler{DB: db}
}

func (h *GenreHandler) List(c *gin.Context) {
	var genres []model.Genre
	if err := h.DB.Order("name ASC").Find(&genres).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": genres})
}
