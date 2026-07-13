package model

import "gorm.io/gorm"

type Notification struct {
        gorm.Model
        UserID  uint   `gorm:"index:idx_notifications_user_read;not null;constraint:OnDelete:CASCADE"`
        Type    string `gorm:"size:50;not null"`
        Title   string `gorm:"size:500;not null"`
        Message string `gorm:"type:text"`
        Link    string `gorm:"size:1000"`
        Read    bool   `gorm:"default:false;index:idx_notifications_user_read"`
        User    User   `gorm:"foreignKey:UserID"`
}
