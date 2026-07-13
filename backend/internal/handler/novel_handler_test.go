package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"readlab/backend/internal/model"
	"readlab/backend/internal/service"
	"readlab/backend/internal/ticket"
)

func setupNovelTest(t *testing.T) (*gin.Engine, string, string) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	if err := db.AutoMigrate(&model.User{}, &model.Novel{}, &model.Genre{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	admin := createTestAdmin(t, db)
	user := createTestUser(t, db, false)

	adminToken, _ := generateTestToken(admin.ID, "test-secret")
	userToken, _ := generateTestToken(user.ID, "test-secret")

	h := NewNovelHandler(db, ticket.NewConfig(db), service.NewNovelService(db))
	r := gin.New()
	r.Use(func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			c.Next()
			return
		}
		tokenStr = tokenStr[7:]
		t, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte("test-secret"), nil
		})
		if err == nil && t.Valid {
			claims := t.Claims.(jwt.MapClaims)
			c.Set("user_id", uint(claims["user_id"].(float64)))
		}
		c.Next()
	})
	api := r.Group("/api/v1")
	protected := api.Group("")
	protected.Use(func(c *gin.Context) {
		_, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	})
	adminGroup := protected.Group("")
	adminGroup.Use(func(c *gin.Context) {
		uid, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		var u model.User
		if err := db.First(&u, uid).Error; err != nil || u.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}
		c.Next()
	})
	{
		adminGroup.POST("/novels", h.Create)
		adminGroup.PUT("/novels/:id", h.Update)
		adminGroup.DELETE("/novels/:id", h.Delete)
	}

	return r, adminToken, userToken
}

func TestNovelCreate_AdminSuccess(t *testing.T) {
	r, adminToken, _ := setupNovelTest(t)

	body, _ := json.Marshal(CreateNovelRequest{
		Title:  "Test Novel",
		Author: "Test Author",
		Status: "ongoing",
	})

	req, _ := http.NewRequest("POST", "/api/v1/novels", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestNovelCreate_NonAdminRejected(t *testing.T) {
	r, _, userToken := setupNovelTest(t)

	body, _ := json.Marshal(CreateNovelRequest{
		Title:  "Test Novel",
		Author: "Test Author",
	})

	req, _ := http.NewRequest("POST", "/api/v1/novels", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+userToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestNovelCreate_UnauthenticatedRejected(t *testing.T) {
	r, _, _ := setupNovelTest(t)

	body, _ := json.Marshal(CreateNovelRequest{
		Title:  "Test Novel",
		Author: "Test Author",
	})

	req, _ := http.NewRequest("POST", "/api/v1/novels", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}
