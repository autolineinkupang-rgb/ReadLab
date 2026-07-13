package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"readlab/backend/internal/model"
	"readlab/backend/internal/ticket"
)

type ShareHandler struct {
	DB     *gorm.DB
	Config *ticket.Config
}

func NewShareHandler(db *gorm.DB, cfg *ticket.Config) *ShareHandler {
	return &ShareHandler{DB: db, Config: cfg}
}

type ShareRequest struct {
	Platform string `json:"platform" binding:"required"`
}

func (h *ShareHandler) Create(c *gin.Context) {
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

	var req ShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing model.Share
	if err := h.DB.Where("user_id = ? AND novel_id = ?", userID, novelID).First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "already shared this novel", "xp_earned": 0})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}
	share := model.Share{
		UserID:   uid,
		NovelID:  uint(novelID),
		Platform: req.Platform,
	}

	if err := h.DB.Create(&share).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record share"})
		return
	}

	var user model.User
	h.DB.First(&user, userID)
	xpAwarded := int64(h.Config.Get("xp_share"))
	if xpAwarded < 1 {
		xpAwarded = 3
	}
	h.DB.Model(&user).Update("xp", gorm.Expr("xp + ?", xpAwarded))

	c.JSON(http.StatusCreated, gin.H{"message": "shared", "xp_earned": xpAwarded})
}
