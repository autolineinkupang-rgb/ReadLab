package model

import "gorm.io/gorm"

type Share struct {
        gorm.Model
        UserID   uint   `gorm:"uniqueIndex:idx_user_novel;not null;constraint:OnDelete:CASCADE"`
        NovelID  uint   `gorm:"uniqueIndex:idx_user_novel;not null;constraint:OnDelete:CASCADE"`
        Platform string `gorm:"size:50;not null"`

        User  User  `gorm:"foreignKey:UserID"`
        Novel Novel `gorm:"foreignKey:NovelID"`
}
