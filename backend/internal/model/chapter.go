package model

import "gorm.io/gorm"

type Chapter struct {
        gorm.Model
        NovelID   uint   `gorm:"not null;index;constraint:OnDelete:CASCADE"`
        Number    int    `gorm:"not null;index"`
        Title     string `gorm:"size:500"`
        Content   string `gorm:"type:text"`
        ContentMD string `gorm:"type:text" json:"content_md"`
        IsLocked  bool   `gorm:"default:false"`
        TicketCost int    `gorm:"default:0"`

        Novel Novel `gorm:"foreignKey:NovelID"`
}
