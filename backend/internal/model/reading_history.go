package model

import "gorm.io/gorm"

type ReadingHistory struct {
	gorm.Model
	UserID    uint `gorm:"uniqueIndex:idx_user_novel_chapter;not null"`
	NovelID   uint `gorm:"uniqueIndex:idx_user_novel_chapter;not null"`
	ChapterID uint `gorm:"uniqueIndex:idx_user_novel_chapter;not null"`
	Progress  float64 `gorm:"default:0"`

	User    User    `gorm:"foreignKey:UserID"`
	Novel   Novel   `gorm:"foreignKey:NovelID"`
	Chapter Chapter `gorm:"foreignKey:ChapterID"`
}
