package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/model"
)

type RequestHandler struct {
	DB *gorm.DB
}

func NewRequestHandler(db *gorm.DB) *RequestHandler {
	return &RequestHandler{DB: db}
}

type CreateRequest struct {
	NovelTitle string `json:"novel_title" binding:"required"`
	NovelURL   string `json:"novel_url"`
	Source     string `json:"source"`
}

func (h *RequestHandler) Create(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	request := model.Request{
		UserID:     userID.(uint),
		NovelTitle: req.NovelTitle,
		NovelURL:   req.NovelURL,
		Source:     req.Source,
		Status:     "pending",
	}

	if err := h.DB.Create(&request).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, request)
}
