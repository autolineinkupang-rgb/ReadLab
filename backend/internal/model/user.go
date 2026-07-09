package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username     string  `gorm:"uniqueIndex;size:100;not null"`
	Email        string  `gorm:"uniqueIndex;size:255;not null"`
	PasswordHash string  `gorm:"size:255;not null" json:"-"`
	DisplayName  string  `gorm:"size:100"`
	AvatarURL    string  `gorm:"size:1000"`
	Tickets        float64    `gorm:"default:0"`
	XP             int64      `gorm:"default:0"`
	Role           string     `gorm:"size:20;default:member"`
	LastDailyClaim *time.Time `gorm:"index"`
	AITranslateProvider    string `gorm:"size:50;default:openrouter"`
	AITranslateModel       string `gorm:"size:100;default:google/gemini-2.0-flash-exp:free"`
	AITranslateKey         string `gorm:"size:500;default:''" json:"-"`
	AITranslateEndpoint    string `gorm:"size:500;default:https://openrouter.ai/api/v1/chat/completions"`
	TranslateTargetLang    string `gorm:"size:10;default:id-ID"`
	AITranslateInstruction string `gorm:"size:2000;default:''"`
}
