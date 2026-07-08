package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/model"
)

type VoteHandler struct {
	DB *gorm.DB
}

func NewVoteHandler(db *gorm.DB) *VoteHandler {
	return &VoteHandler{DB: db}
}

type VoteRequest struct {
	NovelID uint `json:"novel_id" binding:"required"`
}

func (h *VoteHandler) Create(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req VoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vote := model.Vote{
		UserID:  userID.(uint),
		NovelID: req.NovelID,
	}

	if err := h.DB.Create(&vote).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "already voted"})
		return
	}

	h.DB.Model(&model.Novel{}).Where("id = ?", req.NovelID).
		UpdateColumn("votes", gorm.Expr("votes + 1"))

	var user model.User
	h.DB.First(&user, userID)
	xpAwarded := int64(2)
	h.DB.Model(&user).Update("xp", gorm.Expr("xp + ?", xpAwarded))

	c.JSON(http.StatusCreated, gin.H{"message": "voted", "xp_earned": xpAwarded})
}
