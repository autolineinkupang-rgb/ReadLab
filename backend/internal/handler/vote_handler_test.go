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
	"readlab/backend/internal/ticket"
)

func setupVoteTest(t *testing.T) (*gin.Engine, string) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	db.AutoMigrate(&model.User{}, &model.Novel{}, &model.Vote{})

	user := createTestUser(t, db, false)
	token, _ := generateTestToken(user.ID, "test-secret")

	novel := model.Novel{Title: "Test Novel", Slug: "test-novel"}
	db.Create(&novel)

	h := NewVoteHandler(db, ticket.NewConfig(db))
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
	r.POST("/api/v1/votes", h.Create)

	return r, token
}

func TestVote_Success(t *testing.T) {
	r, token := setupVoteTest(t)

	body, _ := json.Marshal(VoteRequest{NovelID: 1})
	req, _ := http.NewRequest("POST", "/api/v1/votes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestVote_DuplicateRejected(t *testing.T) {
	r, token := setupVoteTest(t)

	body, _ := json.Marshal(VoteRequest{NovelID: 1})
	req, _ := http.NewRequest("POST", "/api/v1/votes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	req2, _ := http.NewRequest("POST", "/api/v1/votes", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Errorf("expected 409 for duplicate vote, got %d: %s", w2.Code, w2.Body.String())
	}
}

func TestVote_UnauthenticatedRejected(t *testing.T) {
	r, _ := setupVoteTest(t)

	body, _ := json.Marshal(VoteRequest{NovelID: 1})
	req, _ := http.NewRequest("POST", "/api/v1/votes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}
