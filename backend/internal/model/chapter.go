package model

import "gorm.io/gorm"

type Chapter struct {
	gorm.Model
	NovelID   uint   `gorm:"index;not null"`
	Number    int    `gorm:"not null;index"`
	Title     string `gorm:"size:500"`
	Content   string `gorm:"type:text"`
	IsLocked  bool   `gorm:"default:false"`
	TicketCost int    `gorm:"default:0"`

	Novel Novel `gorm:"foreignKey:NovelID"`
}
