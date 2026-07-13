package service

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"readlab/backend/internal/model"
)

func setupAuthServiceTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	db.AutoMigrate(&model.User{}, &model.TokenBlacklist{})
	return db
}

func newTestAuthService(t *testing.T) (*AuthService, *gorm.DB) {
	db := setupAuthServiceTestDB(t)
	return NewAuthService(db, "test-secret-key", false), db
}

// ── ValidatePassword ──

func TestValidatePassword_Table(t *testing.T) {
	svc, _ := newTestAuthService(t)

	tests := []struct {
		name    string
		pass    string
		wantErr bool
		errSub  string
	}{
		{"too short", "Ab1!", true, "at least 8"},
		{"missing uppercase", "abcdefgh1!", true, "uppercase"},
		{"missing lowercase", "ABCDEFGH1!", true, "lowercase"},
		{"missing digit", "Abcdefgh!!", true, "digit"},
		{"missing special", "Abcdefgh12", true, "special"},
		{"valid password", "Password123!", false, ""},
		{"valid with underscore", "My_Passw0rd!", false, ""},
		{"valid with brackets", "P@ssw0rd[yes]", false, ""},
		{"empty string", "", true, "at least 8"},
		{"exactly 8 chars valid", "Abcdef1!", false, ""},
		{"7 chars too short", "Abcdef1", true, "at least 8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.ValidatePassword(tt.pass)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errSub)
				}
				if tt.errSub != "" && !strings.Contains(err.Error(), tt.errSub) {
					t.Errorf("expected error containing %q, got %q", tt.errSub, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			}
		})
	}
}

func TestValidatePassword_EdgeCases(t *testing.T) {
	svc, _ := newTestAuthService(t)

	// Password with only special chars and digits — no letters
	err := svc.ValidatePassword("12345678!")
	if err == nil || !strings.Contains(err.Error(), "uppercase") {
		t.Errorf("expected uppercase error for no-letter password, got %v", err)
	}

	// Very long valid password
	longPass := "VeryLongPassword123!thatShouldWorkFine"
	err = svc.ValidatePassword(longPass)
	if err != nil {
		t.Errorf("long valid password should pass, got %v", err)
	}
}

// ── HashPassword & CheckPassword ──

func TestHashPassword_ReturnsHash(t *testing.T) {
	svc, _ := newTestAuthService(t)

	hash, err := svc.HashPassword("Password123!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash == "" {
		t.Error("expected non-empty hash")
	}
	if hash == "Password123!" {
		t.Error("hash should not equal plaintext")
	}
}

func TestHashPassword_DifferentEachTime(t *testing.T) {
	svc, _ := newTestAuthService(t)

	hash1, _ := svc.HashPassword("Password123!")
	hash2, _ := svc.HashPassword("Password123!")
	if hash1 == hash2 {
		t.Error("bcrypt should produce different hashes each time due to salt")
	}
}

func TestCheckPassword_CorrectPassword(t *testing.T) {
	svc, _ := newTestAuthService(t)

	hash, _ := svc.HashPassword("Password123!")
	err := svc.CheckPassword(hash, "Password123!")
	if err != nil {
		t.Errorf("expected no error for correct password, got %v", err)
	}
}

func TestCheckPassword_WrongPassword(t *testing.T) {
	svc, _ := newTestAuthService(t)

	hash, _ := svc.HashPassword("Password123!")
	err := svc.CheckPassword(hash, "WrongPassword456!")
	if err == nil {
		t.Error("expected error for wrong password")
	}
}

func TestCheckPassword_InvalidHash(t *testing.T) {
	svc, _ := newTestAuthService(t)

	err := svc.CheckPassword("not-a-valid-bcrypt-hash", "Password123!")
	if err == nil {
		t.Error("expected error for invalid hash")
	}
}

// ── GenerateJTI ──

func TestGenerateJTI_Length(t *testing.T) {
	svc, _ := newTestAuthService(t)

	jti := svc.GenerateJTI()
	if len(jti) != 32 {
		t.Errorf("expected 32-char hex string, got length %d: %q", len(jti), jti)
	}
}

func TestGenerateJTI_IsHexString(t *testing.T) {
	svc, _ := newTestAuthService(t)

	jti := svc.GenerateJTI()
	for _, c := range jti {
		if !(c >= '0' && c <= '9') && !(c >= 'a' && c <= 'f') {
			t.Errorf("JTI contains non-hex character: %c in %q", c, jti)
		}
	}
}

func TestGenerateJTI_Unique(t *testing.T) {
	svc, _ := newTestAuthService(t)

	seen := make(map[string]bool, 100)
	for i := 0; i < 100; i++ {
		jti := svc.GenerateJTI()
		if seen[jti] {
			t.Fatalf("duplicate JTI generated: %s", jti)
		}
		seen[jti] = true
	}
}

// ── GenerateToken ──

func TestGenerateToken_ValidToken(t *testing.T) {
	svc, db := newTestAuthService(t)

	user := model.User{
		Username:     "tokentest",
		Email:        "tokentest@example.com",
		PasswordHash: "$2a$10$abcdefghijklmnopqrstuvwxABCDEFGHIJ",
		DisplayName:  "Token Test",
		Role:         "member",
	}
	db.Create(&user)

	tokenStr, err := svc.GenerateToken(user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("expected non-empty token")
	}

	// Parse the token
	token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("failed to cast claims")
	}

	if claims["user_id"].(float64) != float64(user.ID) {
		t.Errorf("expected user_id %d, got %v", user.ID, claims["user_id"])
	}
	if claims["role"].(string) != "member" {
		t.Errorf("expected role 'member', got '%s'", claims["role"])
	}
	if claims["jti"].(string) == "" {
		t.Error("expected non-empty jti")
	}
}

func TestGenerateToken_NonexistentUser(t *testing.T) {
	svc, _ := newTestAuthService(t)

	_, err := svc.GenerateToken(99999)
	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}
}

// ── BlacklistToken ──

func TestBlacklistToken_ValidToken(t *testing.T) {
	svc, db := newTestAuthService(t)

	user := model.User{
		Username:     "blacklisttest",
		Email:        "blacklist@example.com",
		PasswordHash: "$2a$10$abcdefghijklmnopqrstuvwxABCDEFGHIJ",
		DisplayName:  "Blacklist Test",
		Role:         "member",
	}
	db.Create(&user)

	tokenStr, _ := svc.GenerateToken(user.ID)

	// Extract JTI from token
	token, _, _ := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
	claims := token.Claims.(jwt.MapClaims)
	jti := claims["jti"].(string)

	// Blacklist it
	svc.BlacklistToken(tokenStr)

	// Verify it's in the blacklist
	var bl model.TokenBlacklist
	err := db.Where("jti = ?", jti).First(&bl).Error
	if err != nil {
		t.Fatalf("expected token to be blacklisted, got error: %v", err)
	}
}

func TestBlacklistToken_InvalidToken(t *testing.T) {
	svc, db := newTestAuthService(t)

	// Invalid JWT string — should not panic or error
	svc.BlacklistToken("not.a.valid.token")

	var count int64
	db.Model(&model.TokenBlacklist{}).Count(&count)
	if count != 0 {
		t.Errorf("expected no blacklist entries for invalid token, got %d", count)
	}
}

func TestBlacklistToken_MissingJTI(t *testing.T) {
	svc, db := newTestAuthService(t)

	// Create a valid-looking JWT without jti
	claims := jwt.MapClaims{
		"user_id": float64(1),
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte("test-secret-key"))

	svc.BlacklistToken(tokenStr)

	var count int64
	db.Model(&model.TokenBlacklist{}).Count(&count)
	if count != 0 {
		t.Errorf("expected no blacklist entries for token without jti, got %d", count)
	}
}

func TestBlacklistToken_MissingExp(t *testing.T) {
	svc, db := newTestAuthService(t)

	// Create a JWT with jti but no exp
	claims := jwt.MapClaims{
		"user_id": float64(1),
		"jti":     "test-jti-no-exp",
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte("test-secret-key"))

	svc.BlacklistToken(tokenStr)

	var count int64
	db.Model(&model.TokenBlacklist{}).Count(&count)
	if count != 0 {
		t.Errorf("expected no blacklist entries for token without exp, got %d", count)
	}
}