package handler

import (
	"net/http"
	"strconv"

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

func (h *RequestHandler) List(c *gin.Context) {
	var requests []model.Request
	h.DB.Order("created_at DESC").Find(&requests)

	items := make([]gin.H, len(requests))
	for i, r := range requests {
		items[i] = gin.H{
			"id":          r.ID,
			"novel_title": r.NovelTitle,
			"novel_url":   r.NovelURL,
			"source":      r.Source,
			"status":      r.Status,
			"votes":       r.Votes,
			"created_at":  r.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
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

type ReviewRequest struct {
	Status string `json:"status" binding:"required"`
}

func (h *RequestHandler) Review(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validStatuses := map[string]bool{"approved": true, "rejected": true, "completed": true}
	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status, must be approved/rejected/completed"})
		return
	}

	var request model.Request
	if err := h.DB.First(&request, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "request not found"})
		return
	}

	if err := h.DB.Model(&request).Update("status", req.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, request)
}
