package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"readlab/backend/internal/model"
	"readlab/backend/internal/ticket"
)

type ChapterHandler struct {
	DB *gorm.DB
}

func NewChapterHandler(db *gorm.DB) *ChapterHandler {
	return &ChapterHandler{DB: db}
}

func (h *ChapterHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var chapter model.Chapter
	if err := h.DB.Preload("Novel").First(&chapter, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "chapter not found"})
		return
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
