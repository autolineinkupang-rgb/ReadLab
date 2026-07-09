package model

import "time"

type TokenBlacklist struct {
	ID        uint      `gorm:"primaryKey"`
	JTI       string    `gorm:"uniqueIndex;size:255;not null"`
	ExpiresAt time.Time `gorm:"not null;index"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
