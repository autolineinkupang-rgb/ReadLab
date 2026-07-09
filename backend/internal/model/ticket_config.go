package model

import "gorm.io/gorm"

type TicketConfig struct {
	gorm.Model
	Key   string  `gorm:"uniqueIndex;size:100;not null"`
	Value float64 `gorm:"not null;default:0"`
	Label string  `gorm:"size:255;not null"`
}

func DefaultTicketConfigs() []TicketConfig {
	return []TicketConfig{
		{Key: "daily_reward", Value: 2, Label: "Daily Reward (tickets)"},
		{Key: "novel_contribution", Value: 100, Label: "Novel Contribution Reward (tickets)"},
		{Key: "monthly_leaderboard", Value: 50, Label: "Monthly Leaderboard Reward (tickets)"},
		{Key: "edit_reset_cost", Value: 20, Label: "Edit Limit Reset Cost (tickets)"},
		{Key: "gate_bypass_cost", Value: 50, Label: "Gate Bypass Cost (tickets)"},
		{Key: "replace_review_cost", Value: 100, Label: "Replace Review Cost (tickets)"},
	}
}
