package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/model"
)

type AdminHandler struct {
	DB *gorm.DB
}

func NewAdminHandler(db *gorm.DB) *AdminHandler {
	return &AdminHandler{DB: db}
}

const maxAdmins = 2

func (h *AdminHandler) countAdmins() int64 {
	var count int64
	h.DB.Model(&model.User{}).Where("role = ?", "admin").Count(&count)
	return count
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	role := c.Query("role")
	q := c.Query("q")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	query := h.DB.Model(&model.User{})
	if role != "" {
		query = query.Where("role = ?", role)
	}
	if q != "" {
		like := "%" + q + "%"
		query = query.Where("username ILIKE ? OR email ILIKE ? OR display_name ILIKE ?", like, like, like)
	}

	var total int64
	query.Count(&total)

	var users []model.User
	query.Order("created_at desc").Offset((page - 1) * limit).Limit(limit).Find(&users)

	result := make([]gin.H, len(users))
	for i, u := range users {
		result[i] = gin.H{
			"id":           u.ID,
			"username":     u.Username,
			"email":        u.Email,
			"display_name": u.DisplayName,
			"avatar_url":   u.AvatarURL,
			"role":         u.Role,
			"tickets":      u.Tickets,
			"created_at":   u.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       result,
		"page":       page,
		"limit":      limit,
		"total":      total,
		"total_pages": (int(total) + limit - 1) / limit,
	})
}

func (h *AdminHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var user model.User
	if err := h.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":           user.ID,
		"username":     user.Username,
		"email":        user.Email,
		"display_name": user.DisplayName,
		"avatar_url":   user.AvatarURL,
		"role":         user.Role,
		"tickets":      user.Tickets,
		"created_at":   user.CreatedAt,
	})
}

type UpdateUserRequest struct {
	Role    *string `json:"role"`
	Tickets *float64 `json:"tickets"`
}

func (h *AdminHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var user model.User
	if err := h.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}

	if req.Role != nil {
		validRoles := map[string]bool{"admin": true, "writer": true, "member": true}
		if !validRoles[*req.Role] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
			return
		}

		if *req.Role == "admin" && user.Role != "admin" {
			if h.countAdmins() >= maxAdmins {
				c.JSON(http.StatusConflict, gin.H{"error": "maximum admin limit reached"})
				return
			}
		}

		if user.Role == "admin" && *req.Role != "admin" {
			if h.countAdmins() <= 1 {
				c.JSON(http.StatusConflict, gin.H{"error": "cannot remove last admin"})
				return
			}
		}

		updates["role"] = *req.Role
	}

	if req.Tickets != nil {
		updates["tickets"] = *req.Tickets
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no updates provided"})
		return
	}

	h.DB.Model(&user).Updates(updates)

	c.JSON(http.StatusOK, gin.H{"message": "user updated"})
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if uint(id) == c.GetUint("user_id") {
		c.JSON(http.StatusConflict, gin.H{"error": "cannot delete yourself"})
		return
	}

	var user model.User
	if err := h.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if user.Role == "admin" && h.countAdmins() <= 1 {
		c.JSON(http.StatusConflict, gin.H{"error": "cannot delete last admin"})
		return
	}

	h.DB.Delete(&user)

	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

type CreateAdminRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func (h *AdminHandler) CreateAdmin(c *gin.Context) {
	if h.countAdmins() >= maxAdmins {
		c.JSON(http.StatusConflict, gin.H{"error": "maximum admin limit reached"})
		return
	}

	var req CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user := model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
		DisplayName:  req.Username,
		Role:         "admin",
	}

	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username or email already exists"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
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
	h.DB.Delete(&review)

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

type CreateNewsRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	Type    string `json:"type" binding:"required"`
}

type UpdateNewsRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Type    string `json:"type"`
}

func (h *AdminHandler) CreateNews(c *gin.Context) {
	var req CreateNewsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slug := strings.ToLower(strings.ReplaceAll(req.Title, " ", "-"))
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, slug)
	slug = strings.Trim(slug, "-")

	news := model.News{
		Title:   req.Title,
		Content: req.Content,
		Type:    req.Type,
		Slug:    slug,
	}

	if err := h.DB.Create(&news).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": news})
}

func (h *AdminHandler) UpdateNews(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var news model.News
	if err := h.DB.First(&news, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "news not found"})
		return
	}

	var req UpdateNewsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.Type != "" {
		updates["type"] = req.Type
	}

	if err := h.DB.Model(&news).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.DB.First(&news, id)
	c.JSON(http.StatusOK, gin.H{"data": news})
}

func (h *AdminHandler) DeleteNews(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.DB.Delete(&model.News{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "news deleted"})
}
