package model

import (
	"time"

	"gorm.io/gorm"
)

type PasswordResetToken struct {
	gorm.Model
	UserID    uint      `gorm:"not null;index"`
	Token     string    `gorm:"uniqueIndex;size:255;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	Used      bool      `gorm:"default:false"`
	User      User      `gorm:"foreignKey:UserID"`
}
