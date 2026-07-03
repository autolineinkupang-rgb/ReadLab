package model

import "gorm.io/gorm"

type News struct {
	gorm.Model
	Title   string `gorm:"size:500;not null"`
	Content string `gorm:"type:text"`
	Type    string `gorm:"size:50;default:news;index"` // news, changelog
	Slug    string `gorm:"uniqueIndex;size:500"`
}
