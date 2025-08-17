package models

import (
	"time"
)

// LeadAI represents an AI-managed lead
type LeadAI struct {
	ID           int        `json:"id" db:"id"`
	UserID       string     `json:"user_id" db:"user_id"`
	DeviceID     *string    `json:"device_id" db:"device_id"` // Nullable - assigned during campaign
	Name         string     `json:"name" db:"name"`
	Phone        string     `json:"phone" db:"phone"`
	Email        string     `json:"email" db:"email"`
	Niche        string     `json:"niche" db:"niche"`
	Source       string     `json:"source" db:"source"`
	Status       string     `json:"status" db:"status"`             // pending, assigned, sent, failed
	TargetStatus string     `json:"target_status" db:"target_status"` // prospect/customer
	Notes        string     `json:"notes" db:"notes"`
	AssignedAt   *time.Time `json:"assigned_at" db:"assigned_at"`
	SentAt       *time.Time `json:"sent_at" db:"sent_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// AICampaignProgress tracks the progress of AI campaigns per device
type AICampaignProgress struct {
	ID           int       `json:"id" db:"id"`
	CampaignID   int       `json:"campaign_id" db:"campaign_id"`
	DeviceID     string    `json:"device_id" db:"device_id"`
	LeadsSent    int       `json:"leads_sent" db:"leads_sent"`
	LeadsFailed  int       `json:"leads_failed" db:"leads_failed"`
	Status       string    `json:"status" db:"status"` // active, limit_reached, failed
	LastActivity time.Time `json:"last_activity" db:"last_activity"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
