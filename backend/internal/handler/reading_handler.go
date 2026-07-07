package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/model"
)

type ReadingHandler struct {
	DB *gorm.DB
}

func NewReadingHandler(db *gorm.DB) *ReadingHandler {
	return &ReadingHandler{DB: db}
}

func (h *ReadingHandler) TrackRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	novelID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid novel id"})
		return
	}

	chapterNum, err := strconv.ParseUint(c.Param("num"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chapter number"})
		return
	}

	var chapter model.Chapter
	if err := h.DB.Where("novel_id = ? AND number = ?", novelID, chapterNum).First(&chapter).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "chapter not found"})
		return
	}

	var existing model.ReadingHistory
	result := h.DB.Where("user_id = ? AND novel_id = ? AND chapter_id = ?", userID, novelID, chapter.ID).First(&existing)

	if result.Error != nil {
		entry := model.ReadingHistory{
			UserID:    userID.(uint),
			NovelID:   uint(novelID),
			ChapterID: chapter.ID,
		}
		h.DB.Create(&entry)
	} else {
		h.DB.Model(&existing).Update("progress", 100)
	}

	c.JSON(http.StatusOK, gin.H{"message": "reading progress recorded"})
}

func (h *ReadingHandler) Progress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	novelID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid novel id"})
		return
	}

	var chapterCount int64
	h.DB.Model(&model.ReadingHistory{}).
		Where("user_id = ? AND novel_id = ?", userID, novelID).
		Count(&chapterCount)

	var myReview model.Review
	reviewResult := h.DB.Where("user_id = ? AND novel_id = ?", userID, novelID).Preload("User").First(&myReview)

	var myReviewResp *reviewResponse
	if reviewResult.Error == nil {
		resp := toReviewResponse(myReview)
		myReviewResp = &resp
	}

	c.JSON(http.StatusOK, gin.H{
		"chapter_count": chapterCount,
		"can_review":    chapterCount >= 5,
		"my_review":     myReviewResp,
	})
}
