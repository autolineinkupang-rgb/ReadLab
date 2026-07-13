package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"readlab/backend/internal/model"
	"readlab/backend/internal/ticket"
)

type ReadingHandler struct {
	DB     *gorm.DB
	Config *ticket.Config
}

func NewReadingHandler(db *gorm.DB, cfg *ticket.Config) *ReadingHandler {
	return &ReadingHandler{DB: db, Config: cfg}
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
		var priorReads int64
		h.DB.Model(&model.ReadingHistory{}).Where("user_id = ? AND novel_id = ?", userID, novelID).Count(&priorReads)

		uid, ok := userID.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
			return
		}
		entry := model.ReadingHistory{
			UserID:    uid,
			NovelID:   uint(novelID),
			ChapterID: chapter.ID,
		}
		h.DB.Create(&entry)

		if priorReads == 0 {
			h.DB.Model(&model.Novel{}).Where("id = ?", novelID).UpdateColumn("readers", gorm.Expr("readers + 1"))
		}
	} else {
		h.DB.Model(&existing).Update("progress", 100)
	}

	h.DB.Model(&model.Novel{}).Where("id = ?", novelID).UpdateColumn("views", gorm.Expr("views + 1"))
	c.JSON(http.StatusOK, gin.H{"message": "reading progress recorded"})
}

func (h *ReadingHandler) ClaimXP(c *gin.Context) {
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

	var history model.ReadingHistory
	result := h.DB.Where("user_id = ? AND novel_id = ? AND chapter_id = ?", userID, novelID, chapter.ID).First(&history)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "read the chapter first"})
		return
	}

	if history.XpClaimed {
		c.JSON(http.StatusOK, gin.H{"message": "xp already claimed", "xp_earned": 0})
		return
	}

	xpAwarded := int64(h.Config.Get("xp_read"))
	if xpAwarded < 1 {
		xpAwarded = 10
	}
	tx := h.DB.Begin()
	if err := tx.Model(&model.User{}).Where("id = ?", userID).UpdateColumn("xp", gorm.Expr("xp + ?", xpAwarded)).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to award xp"})
		return
	}
	if err := tx.Model(&history).Update("xp_claimed", true).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark xp"})
		return
	}
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "xp awarded", "xp_earned": xpAwarded})
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

	var lastChapterNumber *int
	var lastChapterTitle string
	var lastRead model.ReadingHistory
	lastReadResult := h.DB.Where("user_id = ? AND novel_id = ?", userID, novelID).
		Preload("Chapter").
		Order("updated_at DESC").
		Limit(1).
		Find(&lastRead)
	if lastReadResult.Error == nil && lastRead.ID != 0 {
		n := lastRead.Chapter.Number
		lastChapterNumber = &n
		lastChapterTitle = lastRead.Chapter.Title
	}

	c.JSON(http.StatusOK, gin.H{
		"chapter_count":      chapterCount,
		"can_review":         chapterCount >= 5,
		"my_review":          myReviewResp,
		"last_chapter":       lastChapterNumber,
		"last_chapter_title": lastChapterTitle,
	})
}
