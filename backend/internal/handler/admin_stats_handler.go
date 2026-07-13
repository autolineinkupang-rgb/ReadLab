package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"readlab/backend/internal/model"
)

type SendTicketsRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

func (h *AdminHandler) updateUserTickets(tx *gorm.DB, userID uint) {
	var sum float64
	tx.Model(&model.TicketUnit{}).
		Where("user_id = ? AND status = 'active'", userID).
		Select("COALESCE(SUM(amount), 0)").Scan(&sum)
	tx.Model(&model.User{}).Where("id = ?", userID).Update("tickets", sum)
}

func (h *AdminHandler) SendTickets(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req SendTicketsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminID := c.GetUint("user_id")

	var adminBalance float64
	h.DB.Model(&model.TicketUnit{}).
		Where("user_id = ? AND status = 'active'", adminID).
		Select("COALESCE(SUM(amount), 0)").Scan(&adminBalance)

	if adminBalance < req.Amount {
		c.JSON(http.StatusConflict, gin.H{"error": "insufficient tickets"})
		return
	}

	var target model.User
	if err := h.DB.First(&target, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		var units []model.TicketUnit
		tx.Where("user_id = ? AND status = 'active'", adminID).
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
					Serial:  model.NewSerial(),
					UserID:  adminID,
					Amount:  excess,
					Status:  "active",
				})
				remaining = 0
			}
		}

		tx.Create(&model.TicketUnit{
			Serial: model.NewSerial(),
			UserID: target.ID,
			Amount: req.Amount,
			Status: "active",
		})

		tx.Create(&model.TicketTransaction{
			UserID:  adminID,
			Amount:  -req.Amount,
			Type:    "spend",
			RefType: "admin_send",
			Note:    "Sent tickets to " + target.Username,
			Date:    now,
		})
		tx.Create(&model.TicketTransaction{
			UserID:  target.ID,
			Amount:  req.Amount,
			Type:    "reward",
			RefType: "admin_gift",
			Note:    "Received tickets from admin",
			Date:    now,
		})

		h.updateUserTickets(tx, adminID)
		h.updateUserTickets(tx, target.ID)

		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send tickets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "tickets sent"})
}

func (h *AdminHandler) GetStats(c *gin.Context) {
	var totalUsers int64
	var totalNovels int64
	var totalChapters int64
	var totalAdmins int64

	h.DB.Model(&model.User{}).Count(&totalUsers)
	h.DB.Model(&model.Novel{}).Count(&totalNovels)
	h.DB.Model(&model.Chapter{}).Count(&totalChapters)
	h.DB.Model(&model.User{}).Where("role = ?", "admin").Count(&totalAdmins)

	c.JSON(http.StatusOK, gin.H{
		"total_users":    totalUsers,
		"total_novels":   totalNovels,
		"total_chapters": totalChapters,
		"total_admins":   totalAdmins,
		"max_admins":     maxAdmins,
	})
}

func (h *AdminHandler) ListReviews(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var total int64
	h.DB.Model(&model.Review{}).Count(&total)

	var reviews []model.Review
	h.DB.Preload("User").Preload("Novel").
		Order("created_at DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&reviews)

	items := make([]gin.H, len(reviews))
	for i, r := range reviews {
		items[i] = gin.H{
			"id":          r.ID,
			"user_id":     r.UserID,
			"username":    r.User.Username,
			"novel_id":    r.NovelID,
			"novel_title": r.Novel.Title,
			"novel_votes": r.Novel.Votes,
			"rating":      r.Rating,
			"content":     r.Content,
			"created_at":  r.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        items,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (int(total) + limit - 1) / limit,
	})
}

func (h *AdminHandler) DeleteReview(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var review model.Review
	if err := h.DB.First(&review, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
		return
	}

	novelID := review.NovelID
	if err := h.DB.Delete(&review).Error; err != nil {
		slog.Error("failed to delete review", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete review"})
		return
	}

	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		var avg float64
		var count int64
		tx.Model(&model.Review{}).
			Select("COALESCE(AVG(rating), 0)").
			Where("novel_id = ?", novelID).
			Scan(&avg)
		tx.Model(&model.Review{}).
			Where("novel_id = ?", novelID).
			Count(&count)
		return tx.Model(&model.Novel{}).Where("id = ?", novelID).
			Updates(map[string]interface{}{
				"Rating":      avg,
				"RatingCount": count,
			}).Error
	}); err != nil {
		slog.Error("failed to update novel rating after review delete", "error", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "review deleted"})
}

func (h *AdminHandler) ListRequests(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	query := h.DB.Model(&model.Request{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var requests []model.Request
	query.Preload("User").
		Order("created_at DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&requests)

	items := make([]gin.H, len(requests))
	for i, r := range requests {
		items[i] = gin.H{
			"id":          r.ID,
			"user_id":     r.UserID,
			"username":    r.User.Username,
			"novel_title": r.NovelTitle,
			"novel_url":   r.NovelURL,
			"source":      r.Source,
			"status":      r.Status,
			"votes":       r.Votes,
			"created_at":  r.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        items,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (int(total) + limit - 1) / limit,
	})
}