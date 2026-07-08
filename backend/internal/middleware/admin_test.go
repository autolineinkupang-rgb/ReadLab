package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/model"
)

func setupAdminTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	db.AutoMigrate(&model.User{})
	return db
}

func TestAdminRequired_AcceptsAdminUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAdminTestDB(t)

	db.Create(&model.User{
		Username:     "admin",
		Email:        "admin@test.com",
		PasswordHash: "hash",
		DisplayName:  "admin",
		Role:         "admin",
	})

	r := gin.New()
	r.GET("/admin", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("role", "admin")
		c.Next()
	}, AdminRequired(db), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminRequired_RejectsNonAdminUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAdminTestDB(t)

	db.Create(&model.User{
		Username:     "user",
		Email:        "user@test.com",
		PasswordHash: "hash",
		DisplayName:  "user",
		Role:         "member",
	})

	r := gin.New()
	r.GET("/admin", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("role", "member")
		c.Next()
	}, AdminRequired(db), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminRequired_RejectsNoRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAdminTestDB(t)

	r := gin.New()
	r.GET("/admin", AdminRequired(db), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminRequired_RejectsNonAdminRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAdminTestDB(t)

	r := gin.New()
	r.GET("/admin", func(c *gin.Context) {
		c.Set("user_id", uint(999))
		c.Set("role", "member")
		c.Next()
	}, AdminRequired(db), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}
