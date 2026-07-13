package model

import (
        "time"

        "gorm.io/gorm"
)

type TicketTransaction struct {
        gorm.Model
        UserID   uint    `gorm:"index;not null;constraint:OnDelete:CASCADE"`
        Amount   float64 `gorm:"not null"`
        Type     string  `gorm:"size:20;not null"` // purchase, spend, reward
        RefType  string  `gorm:"size:50"`
        RefID    uint
        Note     string `gorm:"size:500"`
        Date     time.Time `gorm:"autoCreateTime"`

        User User `gorm:"foreignKey:UserID"`
}
