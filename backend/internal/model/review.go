package model

import "gorm.io/gorm"

type Review struct {
	gorm.Model
	UserID  uint   `gorm:"uniqueIndex:idx_user_novel;not null"`
	NovelID uint   `gorm:"uniqueIndex:idx_user_novel;not null"`
	Rating  uint   `gorm:"not null;check:rating >= 1 AND rating <= 5"`
	Content string `gorm:"type:text;not null"`

	User  User  `gorm:"foreignKey:UserID"`
	Novel Novel `gorm:"foreignKey:NovelID"`
}
