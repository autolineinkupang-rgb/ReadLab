package handler

import (
	"errors"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"readlab/backend/internal/model"
	"readlab/backend/internal/ticket"
)

func (h *NovelHandler) Chapters(c *gin.Context) {
	novelID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 10000 {
		limit = 50
	}

	var total int64
	h.DB.Model(&model.Chapter{}).Where("novel_id = ?", novelID).Count(&total)

	var chapters []model.Chapter
	offset := (page - 1) * limit
	if err := h.DB.Select("id", "novel_id", "number", "title", "is_locked", "ticket_cost", "created_at", "updated_at").
		Where("novel_id = ?", novelID).
		Order("number ASC").
		Offset(offset).Limit(limit).
		Find(&chapters).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  chapters,
		"page":  page,
		"limit": limit,
		"total": total,
	})
}

func (h *NovelHandler) GetChapterByNum(c *gin.Context) {
	novelID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid novel id"})
		return
	}

	num, err := strconv.Atoi(c.Param("num"))
	if err != nil || num < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chapter number"})
		return
	}

	var chapter model.Chapter
	if err := h.DB.Preload("Novel").Where("novel_id = ? AND number = ?", novelID, num).First(&chapter).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "chapter not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if (chapter.Content == "" || chapter.ContentMD == "") &&
		chapter.Novel.SourceURL != "" &&
		strings.Contains(chapter.Novel.SourceURL, "novelfire.net") {
		slug := filepath.Base(chapter.Novel.SourceURL)
		content, err := h.Scraper.ScrapeNovelfireChapterContent(slug, chapter.Number)
		if err == nil && content != "" {
			h.DB.Model(&chapter).Updates(map[string]interface{}{
				"content":    content,
				"content_md": "",
			})
			chapter.Content = content
		}
	}

	if chapter.IsLocked {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "login required to access locked chapter"})
			return
		}

		var user model.User
		if err := h.DB.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}

		var existingTx model.TicketTransaction
		err := h.DB.Where("user_id = ? AND ref_type = ? AND ref_id = ?", user.ID, "chapter", chapter.ID).First(&existingTx).Error
		if err == nil {
			c.JSON(http.StatusOK, chapter)
			return
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		err = h.DB.Transaction(func(tx *gorm.DB) error {
			var sum float64
			tx.Model(&model.TicketUnit{}).
				Where("user_id = ? AND status = 'active'", user.ID).
				Select("COALESCE(SUM(amount), 0)").Scan(&sum)
			if sum < float64(chapter.TicketCost) {
				return ticket.ErrInsufficientTickets
			}

			cost := float64(chapter.TicketCost)
			var units []model.TicketUnit
			tx.Where("user_id = ? AND status = 'active'", user.ID).
				Order("created_at ASC, id ASC").Find(&units)

			remaining := cost
			now := time.Now()
			for _, unit := range units {
				if remaining <= 0 {
					break
				}
				if unit.Amount <= remaining {
					tx.Model(&unit).Updates(map[string]interface{}{
						"status":   "banked",
						"spent_at": &now,
					})
					remaining -= unit.Amount
				} else {
					excess := unit.Amount - remaining
					tx.Model(&unit).Updates(map[string]interface{}{
						"status":   "banked",
						"spent_at": &now,
					})
					tx.Create(&model.TicketUnit{
						Serial: model.NewSerial(),
						UserID: user.ID,
						Amount: excess,
						Status: "active",
					})
					remaining = 0
				}
			}

			tx.Create(&model.TicketTransaction{
				UserID:  user.ID,
				Amount:  -cost,
				Type:    "spend",
				RefType: "chapter",
				RefID:   chapter.ID,
				Date:    now,
				Note:    "Unlock chapter " + strconv.Itoa(chapter.Number) + " of " + chapter.Novel.Title,
			})

			var newSum float64
			tx.Model(&model.TicketUnit{}).
				Where("user_id = ? AND status = 'active'", user.ID).
				Select("COALESCE(SUM(amount), 0)").Scan(&newSum)
			tx.Model(&model.User{}).Where("id = ?", user.ID).Update("tickets", newSum)

			return nil
		})
		if err != nil {
			if errors.Is(err, ticket.ErrInsufficientTickets) {
				c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient tickets"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process ticket payment"})
			return
		}
	}

	c.JSON(http.StatusOK, chapter)
}