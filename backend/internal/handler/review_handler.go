package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/model"
	"wtr-lab-clone/backend/internal/service"
	"wtr-lab-clone/backend/internal/ticket"
)

type ReviewHandler struct {
	DB        *gorm.DB
	Config    *ticket.Config
	ReviewSvc *service.ReviewService
	TicketSvc *service.TicketService
}

func NewReviewHandler(db *gorm.DB, cfg *ticket.Config, reviewSvc *service.ReviewService, ticketSvc *service.TicketService) *ReviewHandler {
	return &ReviewHandler{DB: db, Config: cfg, ReviewSvc: reviewSvc, TicketSvc: ticketSvc}
}

type CreateReviewRequest struct {
	Rating   uint   `json:"rating"`
	Content  string `json:"content" binding:"required,min=1,max=2000"`
	ParentID *uint  `json:"parent_id"`
	Upgrade  bool   `json:"upgrade"`
}

type UpdateReviewRequest struct {
	Rating   uint   `json:"rating" binding:"min=1,max=5"`
	Content  string `json:"content" binding:"required,min=1,max=2000"`
	Upgrade  bool   `json:"upgrade"`
}

type reviewResponse struct {
	ID        uint             `json:"id"`
	Rating    uint             `json:"rating"`
	Content   string           `json:"content"`
	EditCount uint             `json:"edit_count"`
	ParentID  *uint            `json:"parent_id"`
	CreatedAt string           `json:"created_at"`
	User      struct {
		ID          uint   `json:"id"`
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		AvatarURL   string `json:"avatar_url"`
	} `json:"user"`
	Replies []reviewResponse `json:"replies"`
}

func toReviewResponse(r model.Review) reviewResponse {
	var resp reviewResponse
	resp.ID = r.ID
	resp.Rating = r.Rating
	resp.Content = r.Content
	resp.EditCount = r.EditCount
	resp.ParentID = r.ParentID
	resp.CreatedAt = r.CreatedAt.Format("2006-01-02T15:04:05Z")
	resp.User.ID = r.User.ID
	resp.User.Username = r.User.Username
	resp.User.DisplayName = r.User.DisplayName
	resp.User.AvatarURL = r.User.AvatarURL
	if len(r.Replies) > 0 {
		resp.Replies = make([]reviewResponse, len(r.Replies))
		for i, reply := range r.Replies {
			resp.Replies[i] = toReviewResponse(reply)
		}
	} else {
		resp.Replies = []reviewResponse{}
	}
	return resp
}

func (h *ReviewHandler) spendTickets(userID uint, cost float64, refType string, refID uint, note string) error {
	return h.Config.Spend(h.DB, userID, cost, refType, refID, note)
}

func (h *ReviewHandler) List(c *gin.Context) {
	novelID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid novel id"})
		return
	}

	var reviews []model.Review
	if err := h.DB.Where("novel_id = ? AND parent_id IS NULL", novelID).
		Preload("User").
		Preload("Replies", func(db *gorm.DB) *gorm.DB {
			return db.Preload("User").Order("created_at ASC")
		}).
		Order("created_at DESC").
		Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch reviews"})
		return
	}

	var totalCount int64
	h.DB.Model(&model.Review{}).Where("novel_id = ? AND parent_id IS NULL", novelID).Count(&totalCount)

	distribution := map[uint]int64{1: 0, 2: 0, 3: 0, 4: 0, 5: 0}
	type ratingCount struct {
		Rating uint
		Count  int64
	}
	var counts []ratingCount
	h.DB.Model(&model.Review{}).Select("rating, count(*) as count").
		Where("novel_id = ? AND rating > 0", novelID).
		Group("rating").Find(&counts)
	for _, rc := range counts {
		distribution[rc.Rating] = rc.Count
	}

	var avg float64
	if totalCount > 0 {
		h.DB.Model(&model.Review{}).
			Select("COALESCE(avg(rating), 0)").
			Where("novel_id = ? AND rating > 0", novelID).
			Scan(&avg)
	}

	items := make([]reviewResponse, len(reviews))
	for i, r := range reviews {
		items[i] = toReviewResponse(r)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": items,
		"rating_summary": gin.H{
			"average":      avg,
			"count":        totalCount,
			"distribution": distribution,
		},
	})
}

func (h *ReviewHandler) Create(c *gin.Context) {
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

	var novel model.Novel
	if err := h.DB.First(&novel, novelID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "novel not found"})
		return
	}

	var req CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	if req.ParentID != nil {
		var parent model.Review
		if err := h.DB.First(&parent, *req.ParentID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "parent review not found"})
			return
		}
		if parent.ParentID != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot reply to a reply"})
			return
		}
		req.Rating = 0
	} else {
		if req.Rating < 1 || req.Rating > 5 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "rating is required for reviews"})
			return
		}

		var existing model.Review
		hasExisting := h.DB.Where("user_id = ? AND novel_id = ? AND parent_id IS NULL", uid, novelID).First(&existing).Error == nil

		if hasExisting {
			if !req.Upgrade {
					cost := h.Config.Get("replace_review_cost")
					c.JSON(http.StatusConflict, gin.H{
						"error":            "you have already reviewed this novel",
						"upgrade_available": true,
						"upgrade_cost":     cost,
						"upgrade_type":     "duplicate",
					})
					return
				}
				if err := h.spendTickets(uid, h.Config.Get("replace_review_cost"), "upgrade_duplicate", uint(novelID),
				"Replace existing review on novel #"+strconv.Itoa(int(novelID))); err != nil {
				if errors.Is(err, ticket.ErrInsufficientTickets) {
					c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient tickets"})
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process upgrade"})
				}
				return
			}
			h.DB.Delete(&existing)
		} else {
			var chapterCount int64
			h.DB.Model(&model.ReadingHistory{}).
				Where("user_id = ? AND novel_id = ?", uid, novelID).
				Count(&chapterCount)

			if chapterCount < 5 {
				if !req.Upgrade {
					cost := h.Config.Get("gate_bypass_cost")
					c.JSON(http.StatusForbidden, gin.H{
						"error":             "you need to read at least 5 chapters before reviewing",
						"chapter_count":     chapterCount,
						"upgrade_available": true,
						"upgrade_cost":      cost,
						"upgrade_type":      "gate",
					})
					return
				}
				if err := h.spendTickets(uid, h.Config.Get("gate_bypass_cost"), "upgrade_gate", uint(novelID),
					"Bypass 5-chapter review gate for novel #"+strconv.Itoa(int(novelID))); err != nil {
					if errors.Is(err, ticket.ErrInsufficientTickets) {
						c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient tickets"})
					} else {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process upgrade"})
					}
					return
				}
			}
		}
	}

	review := model.Review{
		UserID:   uid,
		NovelID:  uint(novelID),
		Rating:   req.Rating,
		Content:  req.Content,
		ParentID: req.ParentID,
	}

	if err := h.DB.Create(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create review"})
		return
	}

	if req.ParentID == nil {
		if err := h.DB.Transaction(func(tx *gorm.DB) error {
			var avg float64
			var count int64
			tx.Model(&model.Review{}).
				Select("COALESCE(AVG(rating), 0)").
				Where("novel_id = ? AND rating > 0", novelID).
				Scan(&avg)
			tx.Model(&model.Review{}).
				Where("novel_id = ? AND parent_id IS NULL", novelID).
				Count(&count)
			return tx.Model(&model.Novel{}).Where("id = ?", novelID).
				Updates(map[string]interface{}{
					"Rating":      avg,
					"RatingCount": count,
				}).Error
		}); err != nil {
			slog.Error("failed to update novel rating after review", "error", err)
		}
	}

	var user model.User
	h.DB.First(&user, uid)
	xpAwarded := int64(h.Config.Get("xp_review"))
	if xpAwarded < 1 {
		xpAwarded = 5
	}
	h.DB.Model(&user).Update("xp", gorm.Expr("xp + ?", xpAwarded))

	h.DB.Preload("User").First(&review, review.ID)

	c.JSON(http.StatusCreated, gin.H{"data": toReviewResponse(review), "xp_earned": xpAwarded})
}

func (h *ReviewHandler) Update(c *gin.Context) {
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

	reviewID, err := strconv.ParseUint(c.Param("reviewId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}

	var review model.Review
	if err := h.DB.Where("id = ? AND novel_id = ?", reviewID, novelID).First(&review).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	if review.UserID != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only edit your own review"})
		return
	}

	var req UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if review.EditCount >= 5 {
		if !req.Upgrade {
			cost := h.Config.Get("edit_reset_cost")
			c.JSON(http.StatusForbidden, gin.H{
				"error":            "maximum edit limit (5) reached",
				"upgrade_available": true,
				"upgrade_cost":     cost,
				"upgrade_type":     "edit",
			})
			return
		}
		if err := h.spendTickets(uid, h.Config.Get("edit_reset_cost"), "upgrade_edit", review.ID,
			"Upgrade edit limit for review #"+strconv.Itoa(int(review.ID))); err != nil {
			if errors.Is(err, ticket.ErrInsufficientTickets) {
				c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient tickets"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process upgrade"})
			}
			return
		}
		review.EditCount = 0
	}

	oldRating := review.Rating
	review.Rating = req.Rating
	review.Content = req.Content
	review.EditCount++

	if err := h.DB.Save(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update review"})
		return
	}

	if review.ParentID == nil && oldRating != req.Rating {
		if err := h.DB.Transaction(func(tx *gorm.DB) error {
			var avg float64
			var count int64
			tx.Model(&model.Review{}).
				Select("COALESCE(AVG(rating), 0)").
				Where("novel_id = ? AND rating > 0", novelID).
				Scan(&avg)
			tx.Model(&model.Review{}).
				Where("novel_id = ? AND parent_id IS NULL", novelID).
				Count(&count)
			return tx.Model(&model.Novel{}).Where("id = ?", novelID).
				Updates(map[string]interface{}{
					"Rating":      avg,
					"RatingCount": count,
				}).Error
		}); err != nil {
			slog.Error("failed to update novel rating after review update", "error", err)
		}
	}

	h.DB.Preload("User").First(&review, review.ID)

	c.JSON(http.StatusOK, gin.H{"data": toReviewResponse(review)})
}
