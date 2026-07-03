package model

import "gorm.io/gorm"

type Vote struct {
	gorm.Model
	UserID  uint `gorm:"uniqueIndex:idx_user_novel;not null"`
	NovelID uint `gorm:"uniqueIndex:idx_user_novel;not null"`

	User  User  `gorm:"foreignKey:UserID"`
	Novel Novel `gorm:"foreignKey:NovelID"`
}
