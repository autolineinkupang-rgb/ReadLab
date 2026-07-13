package model

import (
        "time"
)

type TicketUnit struct {
        ID        uint       `gorm:"primaryKey"`
        Serial    string     `gorm:"uniqueIndex;size:64;not null"`
        UserID    uint       `gorm:"index;not null;constraint:OnDelete:CASCADE"`
        Amount    float64    `gorm:"not null"`
        Status    string     `gorm:"size:20;default:active;not null"`
        CreatedAt time.Time  `gorm:"autoCreateTime"`
        UpdatedAt time.Time  `gorm:"autoUpdateTime"`
        SpentAt   *time.Time

        User User `gorm:"foreignKey:UserID"`
}
