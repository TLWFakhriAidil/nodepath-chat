package models

import (
	"time"
)

// Campaign represents a marketing campaign
type Campaign struct {
	ID              int       `json:"id" db:"id"`
	UserID          string    `json:"user_id" db:"user_id"`
	DeviceID        string    `json:"device_id" db:"device_id"`
	Title           string    `json:"title" db:"title"`
	Niche           string    `json:"niche" db:"niche"`
	TargetStatus    string    `json:"target_status" db:"target_status"` // prospect, customer, all
	Message         string    `json:"message" db:"message"`
	ImageURL        string    `json:"image_url" db:"image_url"`
	CampaignDate    string    `json:"campaign_date" db:"campaign_date"`
	ScheduledDate   string    `json:"scheduled_date" db:"scheduled_date"`
	TimeSchedule    string    `json:"time_schedule" db:"time_schedule"`
	MinDelaySeconds int       `json:"min_delay_seconds" db:"min_delay_seconds"`
	MaxDelaySeconds int       `json:"max_delay_seconds" db:"max_delay_seconds"`
	Status          string    `json:"status" db:"status"` // pending, sent, failed
	AI              *string   `json:"ai" db:"ai"`         // "ai" for AI campaigns, null for regular
	Limit           int       `json:"limit" db:"\"limit\""` // Device limit for AI campaigns
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}
