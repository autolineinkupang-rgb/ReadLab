package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/model"
)

type SearchHandler struct {
	DB *gorm.DB
}

func NewSearchHandler(db *gorm.DB) *SearchHandler {
	return &SearchHandler{DB: db}
}

func (h *SearchHandler) Search(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var total int64
	h.DB.Model(&model.Novel{}).
		Where("title ILIKE ? OR author ILIKE ?", "%"+q+"%", "%"+q+"%").
		Count(&total)

	var novels []model.Novel
	offset := (page - 1) * limit
	if err := h.DB.Preload("Genres").
		Where("title ILIKE ? OR author ILIKE ?", "%"+q+"%", "%"+q+"%").
		Order("views DESC").
		Offset(offset).Limit(limit).
		Find(&novels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  novels,
		"page":  page,
		"limit": limit,
		"total": total,
	})
}
