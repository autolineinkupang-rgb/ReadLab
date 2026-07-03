package model

import "gorm.io/gorm"

type Genre struct {
	gorm.Model
	Slug string `gorm:"uniqueIndex;size:50;not null"`
	Name string `gorm:"size:100;not null"`
}
