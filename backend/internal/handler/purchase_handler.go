package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/model"
)

type PurchaseHandler struct {
	DB *gorm.DB
}

func NewPurchaseHandler(db *gorm.DB) *PurchaseHandler {
	return &PurchaseHandler{DB: db}
}

type PurchaseRequest struct {
	Amount float64 `json:"amount" binding:"required,min=1,max=1000"`
}

func (h *PurchaseHandler) Create(c *gin.Context) {
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

	var req PurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Mock payment — in production, integrate with payment gateway here
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		tx.Create(&model.TicketUnit{
			Serial: model.NewSerial(),
			UserID: uid,
			Amount: req.Amount,
			Status: "active",
		})

		tx.Create(&model.TicketTransaction{
			UserID: uid,
			Amount: req.Amount,
			Type:   "purchase",
			Date:   time.Now(),
			Note:   "Ticket purchase",
		})

		var sum float64
		tx.Model(&model.TicketUnit{}).
			Where("user_id = ? AND status = 'active'", uid).
			Select("COALESCE(SUM(amount), 0)").Scan(&sum)
		tx.Model(&model.User{}).Where("id = ?", uid).Update("tickets", sum)

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "purchase failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "purchase successful", "amount": req.Amount})
}
