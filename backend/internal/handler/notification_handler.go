package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"readlab/backend/internal/model"
)

type NotificationHandler struct {
	DB *gorm.DB
}

func NewNotificationHandler(db *gorm.DB) *NotificationHandler {
	return &NotificationHandler{DB: db}
}

func (h *NotificationHandler) List(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user identity"})
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
	h.DB.Model(&model.Notification{}).Where("user_id = ?", uid).Count(&total)

	var notifications []model.Notification
	offset := (page - 1) * limit
	if err := h.DB.Where("user_id = ?", uid).Order("created_at DESC").Offset(offset).Limit(limit).Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":         notifications,
		"total":        total,
		"page":         page,
		"limit":        limit,
		"unread_count": h.unreadCount(uid),
	})
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user identity"})
		return
	}

	idStr := c.Param("id")

	if idStr == "all" {
		h.DB.Model(&model.Notification{}).Where("user_id = ? AND read = false", uid).Update("read", true)
	} else {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		h.DB.Model(&model.Notification{}).Where("id = ? AND user_id = ?", id, uid).Update("read", true)
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user identity"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"unread_count": h.unreadCount(uid)})
}

func (h *NotificationHandler) unreadCount(userID uint) int64 {
	var count int64
	h.DB.Model(&model.Notification{}).Where("user_id = ? AND read = false", userID).Count(&count)
	return count
}
