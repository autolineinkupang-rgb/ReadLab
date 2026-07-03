package model

import "gorm.io/gorm"

type Request struct {
	gorm.Model
	UserID      uint   `gorm:"index;not null"`
	NovelTitle  string `gorm:"size:500;not null"`
	NovelURL    string `gorm:"size:1000"`
	Source      string `gorm:"size:100"`
	Status      string `gorm:"size:20;default:pending"`
	Votes       uint   `gorm:"default:0"`

	User User `gorm:"foreignKey:UserID"`
}
