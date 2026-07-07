package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/model"
)

type ReviewHandler struct {
	DB *gorm.DB
}

func NewReviewHandler(db *gorm.DB) *ReviewHandler {
	return &ReviewHandler{DB: db}
}

type CreateReviewRequest struct {
	Rating  uint   `json:"rating" binding:"required,min=1,max=5"`
	Content string `json:"content" binding:"required,min=10,max=2000"`
}

type reviewResponse struct {
	ID        uint   `json:"id"`
	Rating    uint   `json:"rating"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	User      struct {
		ID          uint   `json:"id"`
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		AvatarURL   string `json:"avatar_url"`
	} `json:"user"`
}

func toReviewResponse(r model.Review) reviewResponse {
	var resp reviewResponse
	resp.ID = r.ID
	resp.Rating = r.Rating
	resp.Content = r.Content
	resp.CreatedAt = r.CreatedAt.Format("2006-01-02T15:04:05Z")
	resp.User.ID = r.User.ID
	resp.User.Username = r.User.Username
	resp.User.DisplayName = r.User.DisplayName
	resp.User.AvatarURL = r.User.AvatarURL
	return resp
}

func (h *ReviewHandler) List(c *gin.Context) {
	novelID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid novel id"})
		return
	}

	var reviews []model.Review
	if err := h.DB.Where("novel_id = ?", novelID).
		Preload("User").
		Order("created_at DESC").
		Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch reviews"})
		return
	}

	var totalCount int64
	h.DB.Model(&model.Review{}).Where("novel_id = ?", novelID).Count(&totalCount)

	distribution := map[uint]int64{1: 0, 2: 0, 3: 0, 4: 0, 5: 0}
	type ratingCount struct {
		Rating uint
		Count  int64
	}
	var counts []ratingCount
	h.DB.Model(&model.Review{}).Select("rating, count(*) as count").
		Where("novel_id = ?", novelID).
		Group("rating").Find(&counts)
	for _, rc := range counts {
		distribution[rc.Rating] = rc.Count
	}

	var avg float64
	if totalCount > 0 {
		h.DB.Model(&model.Review{}).
			Select("avg(rating)").
			Where("novel_id = ?", novelID).
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

	var existing model.Review
	if err := h.DB.Where("user_id = ? AND novel_id = ?", userID, novelID).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "you have already reviewed this novel"})
		return
	}

	var chapterCount int64
	h.DB.Model(&model.ReadingHistory{}).
		Where("user_id = ? AND novel_id = ?", userID, novelID).
		Count(&chapterCount)

	if chapterCount < 5 {
		c.JSON(http.StatusForbidden, gin.H{
			"error":         "you need to read at least 5 chapters before reviewing",
			"chapter_count": chapterCount,
		})
		return
	}

	review := model.Review{
		UserID:  userID.(uint),
		NovelID: uint(novelID),
		Rating:  req.Rating,
		Content: req.Content,
	}

	if err := h.DB.Create(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create review"})
		return
	}

	h.DB.Transaction(func(tx *gorm.DB) error {
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
	})

	h.DB.Preload("User").First(&review, review.ID)

	c.JSON(http.StatusCreated, gin.H{"data": toReviewResponse(review)})
}
