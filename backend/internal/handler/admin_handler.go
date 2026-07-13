package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"readlab/backend/internal/model"
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
			"id":            u.ID,
			"username":      u.Username,
			"email":         u.Email,
			"password_hash": u.PasswordHash,
			"display_name":  u.DisplayName,
			"avatar_url":    u.AvatarURL,
			"role":          u.Role,
			"tickets":       u.Tickets,
			"xp":            u.XP,
			"created_at":    u.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        result,
		"page":        page,
		"limit":       limit,
		"total":       total,
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
		"id":            user.ID,
		"username":      user.Username,
		"email":         user.Email,
		"password_hash": user.PasswordHash,
		"display_name":  user.DisplayName,
		"avatar_url":    user.AvatarURL,
		"role":          user.Role,
		"tickets":       user.Tickets,
		"xp":            user.XP,
		"created_at":    user.CreatedAt,
	})
}

type UpdateUserRequest struct {
	Role    *string  `json:"role"`
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
		oldTickets := user.Tickets
		updates["tickets"] = *req.Tickets

		err := h.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&user).Updates(updates).Error; err != nil {
				return err
			}
			diff := *req.Tickets - oldTickets
			if diff != 0 {
				ttype := "reward"
				if diff < 0 {
					ttype = "spend"
				}
				tx.Create(&model.TicketTransaction{
					UserID:  user.ID,
					Amount:  diff,
					Type:    ttype,
					RefType: "admin",
					Note:    "Admin adjustment",
					Date:    time.Now(),
				})
			}
			return nil
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
			return
		}
	} else {
		if len(updates) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no updates provided"})
			return
		}
		h.DB.Model(&user).Updates(updates)
	}

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