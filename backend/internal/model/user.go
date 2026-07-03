package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username     string `gorm:"uniqueIndex;size:100;not null"`
	Email        string `gorm:"uniqueIndex;size:255;not null"`
	PasswordHash string `gorm:"size:255;not null"`
	DisplayName  string `gorm:"size:100"`
	AvatarURL    string `gorm:"size:1000"`
	Tickets      float64 `gorm:"default:0"`
	IsAdmin      bool    `gorm:"default:false"`
}

type Session struct {
	gorm.Model
	UserID    uint      `gorm:"index;not null"`
	Token     string    `gorm:"uniqueIndex;size:500;not null"`
	ExpiresAt time.Time `gorm:"not null"`

	User User `gorm:"foreignKey:UserID"`
}
