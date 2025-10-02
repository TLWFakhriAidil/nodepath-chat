package models

import (
	"time"
)

// AISettings represents AI configuration settings for a device
type AISettings struct {
	ID             string    `json:"id" db:"id"`
	IDDevice       string    `json:"id_device" db:"id_device"`
	SystemPrompt   string    `json:"system_prompt" db:"system_prompt"`
	ClosingPrompt  string    `json:"closing_prompt" db:"closing_prompt"`
	InstancePrompt string    `json:"instance_prompt" db:"instance_prompt"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}
