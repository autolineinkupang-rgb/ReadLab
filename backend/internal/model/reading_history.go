package model

import "gorm.io/gorm"

type ReadingHistory struct {
        gorm.Model
        UserID    uint `gorm:"uniqueIndex:idx_user_novel_chapter;not null;constraint:OnDelete:CASCADE"`
        NovelID   uint `gorm:"uniqueIndex:idx_user_novel_chapter;not null;constraint:OnDelete:CASCADE"`
        ChapterID uint `gorm:"uniqueIndex:idx_user_novel_chapter;not null;constraint:OnDelete:CASCADE"`
        Progress  float64 `gorm:"default:0"`
        XpClaimed bool    `gorm:"default:false"`

        User    User    `gorm:"foreignKey:UserID"`
        Novel   Novel   `gorm:"foreignKey:NovelID"`
        Chapter Chapter `gorm:"foreignKey:ChapterID"`
}
