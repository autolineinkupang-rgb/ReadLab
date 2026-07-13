package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"readlab/backend/internal/model"
)

type ClaimBankRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

func (h *AdminHandler) BankBalance(c *gin.Context) {
	var sum float64
	h.DB.Model(&model.TicketUnit{}).
		Where("status = 'banked'").
		Select("COALESCE(SUM(amount), 0)").Scan(&sum)

	var count int64
	h.DB.Model(&model.TicketUnit{}).
		Where("status = 'banked'").
		Count(&count)

	c.JSON(http.StatusOK, gin.H{
		"balance": sum,
		"units":   count,
	})
}

func (h *AdminHandler) BankClaim(c *gin.Context) {
	var req ClaimBankRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminID := c.GetUint("user_id")

	var bankSum float64
	h.DB.Model(&model.TicketUnit{}).
		Where("status = 'banked'").
		Select("COALESCE(SUM(amount), 0)").Scan(&bankSum)

	if bankSum < req.Amount {
		c.JSON(http.StatusConflict, gin.H{"error": "insufficient bank balance"})
		return
	}

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		var units []model.TicketUnit
		tx.Where("status = 'banked'").
			Order("created_at ASC, id ASC").
			Find(&units)

		remaining := req.Amount
		now := time.Now()

		for _, unit := range units {
			if remaining <= 0 {
				break
			}
			if unit.Amount <= remaining {
				tx.Model(&unit).Updates(map[string]interface{}{
					"status":   "spent",
					"spent_at": &now,
				})
				remaining -= unit.Amount
			} else {
				excess := unit.Amount - remaining
				tx.Model(&unit).Updates(map[string]interface{}{
					"status":   "spent",
					"spent_at": &now,
				})
				tx.Create(&model.TicketUnit{
					Serial: model.NewSerial(),
					UserID: adminID,
					Amount: excess,
					Status: "banked",
				})
				remaining = 0
			}
		}

		tx.Create(&model.TicketUnit{
			Serial: model.NewSerial(),
			UserID: adminID,
			Amount: req.Amount,
			Status: "active",
		})

		tx.Create(&model.TicketTransaction{
			UserID:  adminID,
			Amount:  req.Amount,
			Type:    "reward",
			RefType: "bank_claim",
			Note:    "Claimed from ticket bank",
			Date:    time.Now(),
		})

		h.updateUserTickets(tx, adminID)

		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to claim tickets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "tickets claimed from bank"})
}