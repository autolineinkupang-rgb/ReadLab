package handler

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"readlab/backend/internal/model"
)

var testUserCounter int

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func bcryptHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func createTestUser(t *testing.T, db *gorm.DB, isAdmin bool) model.User {
	testUserCounter++
	username := fmt.Sprintf("testuser%d", testUserCounter)
	email := fmt.Sprintf("test%d@example.com", testUserCounter)
	hash, _ := bcryptHash("password123")
	role := "member"
	if isAdmin {
		role = "admin"
	}
	user := model.User{
		Username:     username,
		Email:        email,
		PasswordHash: hash,
		DisplayName:  username,
		Role:         role,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	return user
}

func createTestAdmin(t *testing.T, db *gorm.DB) model.User {
	return createTestUser(t, db, true)
}

func generateTestToken(userID uint, secret string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
