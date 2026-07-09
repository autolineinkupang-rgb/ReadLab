package model

import "gorm.io/gorm"

type Tag struct {
	gorm.Model
	Name string `gorm:"size:100;uniqueIndex;not null"`
	Slug string `gorm:"uniqueIndex;size:100;not null"`
}

type NovelTag struct {
	NovelID uint `gorm:"primaryKey"`
	TagID   uint `gorm:"primaryKey"`
}
