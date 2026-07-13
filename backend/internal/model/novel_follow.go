package model

import "gorm.io/gorm"

type NovelFollow struct {
        gorm.Model
        UserID  uint `gorm:"uniqueIndex:idx_user_novel_follow;not null;constraint:OnDelete:CASCADE"`
        NovelID uint `gorm:"uniqueIndex:idx_user_novel_follow;not null;constraint:OnDelete:CASCADE"`

        User  User  `gorm:"foreignKey:UserID"`
        Novel Novel `gorm:"foreignKey:NovelID"`
}
