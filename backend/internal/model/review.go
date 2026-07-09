package model

import "gorm.io/gorm"

type Review struct {
	gorm.Model
	UserID    uint   `gorm:"uniqueIndex:idx_user_novel_parent;not null"`
	NovelID   uint   `gorm:"uniqueIndex:idx_user_novel_parent;not null"`
	Rating  uint   `gorm:"not null;check:rating >= 0 AND rating <= 5"`
	Content   string `gorm:"type:text;not null"`
	EditCount uint   `gorm:"default:0"`
	ParentID  *uint  `gorm:"uniqueIndex:idx_user_novel_parent;index"`

	User    User     `gorm:"foreignKey:UserID"`
	Novel   Novel    `gorm:"foreignKey:NovelID"`
	Replies []Review `gorm:"foreignKey:ParentID"`
}
