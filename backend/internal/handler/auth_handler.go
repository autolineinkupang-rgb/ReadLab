package handler

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/model"
)

type AuthHandler struct {
	DB           *gorm.DB
	JWTSecret    string
	CookieSecure bool
}

func NewAuthHandler(db *gorm.DB, jwtSecret string, cookieSecure bool) *AuthHandler {
	return &AuthHandler{DB: db, JWTSecret: jwtSecret, CookieSecure: cookieSecure}
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validatePassword(req.Password); err != nil {
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
		Role:         "member",
	}

	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username or email already exists"})
		return
	}

	token, err := h.generateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	h.setCookie(c, token)

	c.JSON(http.StatusCreated, gin.H{
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user model.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := h.generateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	h.setCookie(c, token)

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var user model.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"display_name": user.DisplayName,
		"avatar_url": user.AvatarURL,
		"tickets":    user.Tickets,
		"role":       user.Role,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.CookieSecure,
		SameSite: http.SameSiteStrictMode,
	})
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (h *AuthHandler) generateToken(userID uint) (string, error) {
	var user model.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    user.Role,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.JWTSecret))
}

func (h *AuthHandler) setCookie(c *gin.Context, token string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		MaxAge:   7 * 24 * 3600,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.CookieSecure,
		SameSite: http.SameSiteStrictMode,
	})
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString
	hasLower := regexp.MustCompile(`[a-z]`).MatchString
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~]`).MatchString

	if !hasUpper(password) {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower(password) {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasDigit(password) {
		return fmt.Errorf("password must contain at least one digit")
	}
	if !hasSpecial(password) {
		return fmt.Errorf("password must contain at least one special character")
	}
	return nil
}
