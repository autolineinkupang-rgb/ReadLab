package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/model"
)

type AdminChapterHandler struct {
	DB *gorm.DB
}

func NewAdminChapterHandler(db *gorm.DB) *AdminChapterHandler {
	return &AdminChapterHandler{DB: db}
}

type CreateChapterRequest struct {
	Number     int    `json:"number"`
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content" binding:"required"`
	IsLocked   bool   `json:"is_locked"`
	TicketCost int    `json:"ticket_cost"`
}

type UpdateChapterRequest struct {
	Title      *string `json:"title"`
	Content    *string `json:"content"`
	IsLocked   *bool   `json:"is_locked"`
	TicketCost *int    `json:"ticket_cost"`
}

type chapterResponse struct {
	ID         uint   `json:"id"`
	NovelID    uint   `json:"novel_id"`
	Number     int    `json:"number"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	IsLocked   bool   `json:"is_locked"`
	TicketCost int    `json:"ticket_cost"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

func toChapterResponse(c model.Chapter) chapterResponse {
	return chapterResponse{
		ID:         c.ID,
		NovelID:    c.NovelID,
		Number:     c.Number,
		Title:      c.Title,
		Content:    c.Content,
		IsLocked:   c.IsLocked,
		TicketCost: c.TicketCost,
		CreatedAt:  c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (h *AdminChapterHandler) Create(c *gin.Context) {
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

	role, _ := c.Get("role")
	if role != "admin" {
		var novel model.Novel
		if err := h.DB.First(&novel, novelID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "novel not found"})
			return
		}
		uid, ok := userID.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
			return
		}
		if novel.WriterID == nil || *novel.WriterID != uid {
			c.JSON(http.StatusForbidden, gin.H{"error": "you do not own this novel"})
			return
		}
	}

	var req CreateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Number == 0 {
		var maxNum int
		h.DB.Model(&model.Chapter{}).
			Select("COALESCE(MAX(number), 0)").
			Where("novel_id = ?", novelID).
			Scan(&maxNum)
		req.Number = maxNum + 1
	}

	var existing model.Chapter
	if err := h.DB.Where("novel_id = ? AND number = ?", novelID, req.Number).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "chapter number already exists"})
		return
	}

	chapter := model.Chapter{
		NovelID:    uint(novelID),
		Number:     req.Number,
		Title:      req.Title,
		Content:    req.Content,
		IsLocked:   req.IsLocked,
		TicketCost: req.TicketCost,
	}

	if err := h.DB.Create(&chapter).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create chapter"})
		return
	}

	h.DB.Model(&model.Novel{}).Where("id = ?", novelID).UpdateColumn("chapters", gorm.Expr("chapters + 1"))
	h.DB.First(&chapter, chapter.ID)

	c.JSON(http.StatusCreated, gin.H{"chapter": toChapterResponse(chapter)})
}

func (h *AdminChapterHandler) Update(c *gin.Context) {
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

	chapterID, err := strconv.ParseUint(c.Param("chapterID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chapter id"})
		return
	}

	role, _ := c.Get("role")
	if role != "admin" {
		var novel model.Novel
		if err := h.DB.First(&novel, novelID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "novel not found"})
			return
		}
		uid, ok := userID.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
			return
		}
		if novel.WriterID == nil || *novel.WriterID != uid {
			c.JSON(http.StatusForbidden, gin.H{"error": "you do not own this novel"})
			return
		}
	}

	var chapter model.Chapter
	if err := h.DB.Where("id = ? AND novel_id = ?", chapterID, novelID).First(&chapter).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "chapter not found"})
		return
	}

	var req UpdateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.IsLocked != nil {
		updates["is_locked"] = *req.IsLocked
	}
	if req.TicketCost != nil {
		updates["ticket_cost"] = *req.TicketCost
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no updates provided"})
		return
	}

	h.DB.Model(&chapter).Updates(updates)
	h.DB.First(&chapter, chapter.ID)

	c.JSON(http.StatusOK, gin.H{"chapter": toChapterResponse(chapter)})
}

func (h *AdminChapterHandler) Delete(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chapter id"})
		return
	}

	var chapter model.Chapter
	if err := h.DB.First(&chapter, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "chapter not found"})
		return
	}

	role, _ := c.Get("role")
	if role != "admin" {
		var novel model.Novel
		if err := h.DB.First(&novel, chapter.NovelID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "novel not found"})
			return
		}
		uid, ok := userID.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
			return
		}
		if novel.WriterID == nil || *novel.WriterID != uid {
			c.JSON(http.StatusForbidden, gin.H{"error": "you do not own this novel"})
			return
		}
	}

	h.DB.Delete(&chapter)
	h.DB.Model(&model.Novel{}).Where("id = ?", chapter.NovelID).UpdateColumn("chapters", gorm.Expr("GREATEST(chapters - 1, 0)"))

	c.JSON(http.StatusOK, gin.H{"message": "chapter deleted"})
}

func (h *AdminChapterHandler) List(c *gin.Context) {
	novelID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid novel id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var total int64
	h.DB.Model(&model.Chapter{}).Where("novel_id = ?", novelID).Count(&total)

	var chapters []model.Chapter
	h.DB.Where("novel_id = ?", novelID).
		Order("number ASC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&chapters)

	items := make([]chapterResponse, len(chapters))
	for i, c := range chapters {
		items[i] = toChapterResponse(c)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        items,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (int(total) + limit - 1) / limit,
	})
}

func (h *AdminChapterHandler) Get(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chapter id"})
		return
	}

	var chapter model.Chapter
	if err := h.DB.First(&chapter, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "chapter not found"})
		return
	}

	role, _ := c.Get("role")
	if role != "admin" {
		var novel model.Novel
		if err := h.DB.First(&novel, chapter.NovelID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "novel not found"})
			return
		}
		uid, ok := userID.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
			return
		}
		if novel.WriterID == nil || *novel.WriterID != uid {
			c.JSON(http.StatusForbidden, gin.H{"error": "you do not own this novel"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"chapter": toChapterResponse(chapter)})
}
