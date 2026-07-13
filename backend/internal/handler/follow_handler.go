package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"readlab/backend/internal/model"
)

type FollowHandler struct {
	DB *gorm.DB
}

func NewFollowHandler(db *gorm.DB) *FollowHandler {
	return &FollowHandler{DB: db}
}

func (h *FollowHandler) Create(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	novelID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid novel id"})
		return
	}

	var novel model.Novel
	if err := h.DB.First(&novel, novelID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "novel not found"})
		return
	}

	follow := model.NovelFollow{
		UserID:  uid,
		NovelID: uint(novelID),
	}

	if err := h.DB.Create(&follow).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "already following this novel"})
		return
	}

	h.DB.Create(&model.Notification{
		UserID:  uid,
		Type:    "follow",
		Title:   "Novel Followed",
		Message: "You started following " + novel.Title,
		Link:    "/en/novel/" + strconv.FormatUint(novelID, 10) + "/" + novel.Slug,
	})

	c.JSON(http.StatusCreated, gin.H{"message": "now following this novel"})
}

func (h *FollowHandler) Delete(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	novelID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid novel id"})
		return
	}

	result := h.DB.Where("user_id = ? AND novel_id = ?", uid, novelID).Delete(&model.NovelFollow{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not following this novel"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "unfollowed"})
}

func (h *FollowHandler) Check(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	novelID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid novel id"})
		return
	}

	var count int64
	h.DB.Model(&model.NovelFollow{}).Where("user_id = ? AND novel_id = ?", uid, novelID).Count(&count)

	c.JSON(http.StatusOK, gin.H{"following": count > 0})
}
