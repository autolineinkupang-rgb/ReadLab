package model

import "gorm.io/gorm"

type NovelFollow struct {
	gorm.Model
	UserID  uint `gorm:"uniqueIndex:idx_user_novel_follow;not null"`
	NovelID uint `gorm:"uniqueIndex:idx_user_novel_follow;not null"`

	User  User  `gorm:"foreignKey:UserID"`
	Novel Novel `gorm:"foreignKey:NovelID"`
}
