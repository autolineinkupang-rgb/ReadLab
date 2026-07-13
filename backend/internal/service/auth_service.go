package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"readlab/backend/internal/model"
)

type AuthService struct {
	DB           *gorm.DB
	JWTSecret    string
	CookieSecure bool
}

func NewAuthService(db *gorm.DB, jwtSecret string, cookieSecure bool) *AuthService {
	return &AuthService{DB: db, JWTSecret: jwtSecret, CookieSecure: cookieSecure}
}

func (s *AuthService) ValidatePassword(password string) error {
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

func (s *AuthService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func (s *AuthService) CheckPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (s *AuthService) GenerateJTI() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *AuthService) GenerateToken(userID uint) (string, error) {
	var user model.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    user.Role,
		"jti":     s.GenerateJTI(),
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.JWTSecret))
}

func (s *AuthService) BlacklistToken(tokenStr string) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return
	}
	jti, ok := claims["jti"].(string)
	if !ok || jti == "" {
		return
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		return
	}
	s.DB.Create(&model.TokenBlacklist{
		JTI:       jti,
		ExpiresAt: time.Unix(int64(exp), 0),
	})
}
