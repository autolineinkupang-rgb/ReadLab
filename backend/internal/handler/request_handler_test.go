package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"wtr-lab-clone/backend/internal/model"
)

func setupRequestTest(t *testing.T) (*gin.Engine, string, string) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	db.AutoMigrate(&model.User{}, &model.Request{})

	admin := createTestAdmin(t, db)
	user := createTestUser(t, db, false)

	adminToken, _ := generateTestToken(admin.ID, "test-secret")
	userToken, _ := generateTestToken(user.ID, "test-secret")

	db.Create(&model.Request{
		UserID:     user.ID,
		NovelTitle: "Test Request",
		Status:     "pending",
	})

	h := NewRequestHandler(db)
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
		adminGroup.PUT("/requests/:id", h.Review)
	}
	protected.POST("/requests", h.Create)

	return r, adminToken, userToken
}

func TestRequestReview_AdminSuccess(t *testing.T) {
	r, adminToken, _ := setupRequestTest(t)

	body, _ := json.Marshal(ReviewRequest{Status: "approved"})
	req, _ := http.NewRequest("PUT", "/api/v1/requests/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRequestReview_NonAdminRejected(t *testing.T) {
	r, _, userToken := setupRequestTest(t)

	body, _ := json.Marshal(ReviewRequest{Status: "approved"})
	req, _ := http.NewRequest("PUT", "/api/v1/requests/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+userToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRequestReview_UnauthenticatedRejected(t *testing.T) {
	r, _, _ := setupRequestTest(t)

	body, _ := json.Marshal(ReviewRequest{Status: "approved"})
	req, _ := http.NewRequest("PUT", "/api/v1/requests/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}
