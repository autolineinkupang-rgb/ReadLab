package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"readlab/backend/internal/model"
	"readlab/backend/internal/service"
)

func setupAuthTest(t *testing.T) (*gorm.DB, *AuthHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	db.AutoMigrate(&model.User{})

	authSvc := service.NewAuthService(db, "test-secret", false)
	h := NewAuthHandler(db, "test-secret", false, nil, authSvc)
	r := gin.New()
	api := r.Group("/api/v1")
	api.POST("/auth/register", h.Register)
	api.POST("/auth/login", h.Login)

	return db, h, r
}

func TestAuthRegister_Success(t *testing.T) {
	_, _, r := setupAuthTest(t)

	body, _ := json.Marshal(RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "Password123!",
	})

	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuthRegister_DuplicateEmail(t *testing.T) {
	db, _, r := setupAuthTest(t)

	db.Create(&model.User{
		Username:     "existing",
		Email:        "test@example.com",
		PasswordHash: "hash",
		DisplayName:  "existing",
	})

	body, _ := json.Marshal(RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "Password123!",
	})

	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuthLogin_Success(t *testing.T) {
	db, _, r := setupAuthTest(t)

	hash, _ := bcryptHash("Password123!")
	db.Create(&model.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hash,
		DisplayName:  "testuser",
	})

	body, _ := json.Marshal(LoginRequest{
		Email:    "test@example.com",
		Password: "Password123!",
	})

	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuthLogin_WrongPassword(t *testing.T) {
	db, _, r := setupAuthTest(t)

	hash, _ := bcryptHash("correctpassword")
	db.Create(&model.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hash,
		DisplayName:  "testuser",
	})

	body, _ := json.Marshal(LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	})

	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}
