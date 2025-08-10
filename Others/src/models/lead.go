package models

import (
	"time"
)

// Lead represents a lead/contact
type Lead struct {
	ID           string    `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	DeviceID     string    `json:"device_id" db:"device_id"`
	Name         string    `json:"name" db:"name"`
	Phone        string    `json:"phone" db:"phone"`
	Email        string    `json:"email" db:"email"`
	Niche        string    `json:"niche" db:"niche"` // For matching with campaigns/sequences
	Source       string    `json:"source" db:"source"`
	Status       string    `json:"status" db:"status"` // Keep for backward compatibility
	TargetStatus string    `json:"target_status" db:"target_status"` // New column: prospect/customer
	Trigger      string    `json:"trigger" db:"trigger"` // New column: comma-separated sequence triggers
	Notes        string    `json:"notes" db:"notes"`
	Platform     string    `json:"platform" db:"platform"` // New column: Whacenter, etc.
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}