package model

import (
	"time"

	"gorm.io/gorm"
)

type Novel struct {
	gorm.Model
	Title       string  `gorm:"size:500;not null;index"`
	AltTitle    string  `gorm:"size:500"`
	Slug        string  `gorm:"uniqueIndex;size:500;not null"`
	Author      string  `gorm:"size:200"`
	AuthorSlug  string  `gorm:"size:500"`
	Status      string  `gorm:"size:20;default:ongoing;index"`
	Views       uint64  `gorm:"default:0"`
	Rating      float64 `gorm:"default:0"`
	RatingCount uint    `gorm:"default:0"`
	Chapters    int     `gorm:"default:0"`
	Readers     int     `gorm:"default:0"`
	Chars       string  `gorm:"size:20"`
	AIPercent   string  `gorm:"size:10"`
	Description string  `gorm:"type:text"`
	CoverURL    string  `gorm:"size:1000"`
	SourceURL   string  `gorm:"size:1000"`
	RequestedBy string  `gorm:"size:200"`
	ReleasedBy  string  `gorm:"size:200"`
	AddedAt     time.Time `gorm:"autoCreateTime"`

	Genres []Genre `gorm:"many2many:novel_genres;"`
}

type NovelGenre struct {
	NovelID uint `gorm:"primaryKey"`
	GenreID uint `gorm:"primaryKey"`
}
