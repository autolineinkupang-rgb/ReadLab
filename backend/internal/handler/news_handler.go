package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/model"
)

type NewsHandler struct {
	DB *gorm.DB
}

func NewNewsHandler(db *gorm.DB) *NewsHandler {
	return &NewsHandler{DB: db}
}

func (h *NewsHandler) List(c *gin.Context) {
	newsType := c.Query("type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}

	query := h.DB.Model(&model.News{})
	if newsType != "" {
		query = query.Where("type = ?", newsType)
	}

	var total int64
	query.Count(&total)

	var news []model.News
	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&news).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  news,
		"page":  page,
		"limit": limit,
		"total": total,
	})
}

func (h *NewsHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		slug := c.Param("id")
		var news model.News
		if err := h.DB.Where("slug = ?", slug).First(&news).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "news not found"})
			return
		}
		c.JSON(http.StatusOK, news)
		return
	}

	var news model.News
	if err := h.DB.First(&news, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "news not found"})
		return
	}

	c.JSON(http.StatusOK, news)
}
