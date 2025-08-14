package models

import (
	"time"
)

// Sequence model for drip campaigns - simplified like campaigns
type Sequence struct {
	ID              string         `json:"id" db:"id"`
	UserID          string         `json:"user_id" db:"user_id"`
	DeviceID        *string        `json:"device_id" db:"device_id"` // Nullable - sequences use all user devices
	Name            string         `json:"name" db:"name"`
	Description     string         `json:"description" db:"description"`
	Niche           string         `json:"niche" db:"niche"`
	TargetStatus    string         `json:"target_status" db:"target_status"` // prospect, customer, all
	Status          string         `json:"status" db:"status"` // draft, active, paused
	Trigger         string         `json:"trigger" db:"trigger"` // Main trigger for the sequence
	StartTrigger    string         `json:"start_trigger" db:"start_trigger"` // Deprecated - kept for backward compatibility
	EndTrigger      string         `json:"end_trigger" db:"end_trigger"` // Deprecated - kept for backward compatibility
	TotalDays       int            `json:"total_days" db:"total_days"` // Added
	IsActive        bool           `json:"is_active" db:"is_active"`   // Added
	TimeSchedule    string         `json:"time_schedule" db:"schedule_time"`
	MinDelaySeconds int            `json:"min_delay_seconds" db:"min_delay_seconds"`
	MaxDelaySeconds int            `json:"max_delay_seconds" db:"max_delay_seconds"`
	ContactsCount   int            `json:"contacts_count" db:"contacts_count"`
	// Progress tracking fields
	TotalContacts      int            `json:"total_contacts" db:"total_contacts"`
	ActiveContacts     int            `json:"active_contacts" db:"active_contacts"`
	CompletedContacts  int            `json:"completed_contacts" db:"completed_contacts"`
	FailedContacts     int            `json:"failed_contacts" db:"failed_contacts"`
	ProgressPercentage float64        `json:"progress_percentage" db:"progress_percentage"`
	LastActivityAt     *time.Time     `json:"last_activity_at" db:"last_activity_at"`
	EstimatedCompletionAt *time.Time  `json:"estimated_completion_at" db:"estimated_completion_at"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
	Steps           []SequenceStep `json:"steps,omitempty"`
}

// SequenceStep model for each day in sequence - simplified
type SequenceStep struct {
	ID               string    `json:"id" db:"id"`
	SequenceID       string    `json:"sequence_id" db:"sequence_id"`
	DayNumber        int       `json:"day_number" db:"day_number"`
	Trigger          string    `json:"trigger" db:"trigger"`
	NextTrigger      string    `json:"next_trigger" db:"next_trigger"` // Next step trigger
	TriggerDelayHours int      `json:"trigger_delay_hours" db:"trigger_delay_hours"` // Hours to wait
	IsEntryPoint     bool      `json:"is_entry_point" db:"is_entry_point"` // Sequence start
	MessageType      string    `json:"message_type" db:"message_type"`
	TimeSchedule     string    `json:"time_schedule" db:"time_schedule"`
	Content          string    `json:"content" db:"content"`
	MediaURL         string    `json:"media_url" db:"media_url"`
	Caption          string    `json:"caption" db:"caption"`
	MinDelaySeconds  int       `json:"min_delay_seconds" db:"min_delay_seconds"`
	MaxDelaySeconds  int       `json:"max_delay_seconds" db:"max_delay_seconds"`
	DelayDays        int       `json:"delay_days" db:"delay_days"`
}

// SequenceContact model for tracking individual progress
type SequenceContact struct {
	ID                   string     `json:"id" db:"id"`
	SequenceID           string     `json:"sequence_id" db:"sequence_id"`
	ContactPhone         string     `json:"contact_phone" db:"contact_phone"`
	ContactName          string     `json:"contact_name" db:"contact_name"`
	CurrentStep          int        `json:"current_step" db:"current_step"`
	Status               string     `json:"status" db:"status"` // active, completed, paused
	CompletedAt          *time.Time `json:"completed_at" db:"completed_at"`
	CurrentTrigger       string     `json:"current_trigger" db:"current_trigger"`
	NextTriggerTime      *time.Time `json:"next_trigger_time" db:"next_trigger_time"`
	ProcessingDeviceID   *string    `json:"processing_device_id" db:"processing_device_id"`
	LastError            *string    `json:"last_error" db:"last_error"`
	RetryCount           int        `json:"retry_count" db:"retry_count"`
	AssignedDeviceID     *string    `json:"assigned_device_id" db:"assigned_device_id"`
	ProcessingStartedAt  *time.Time `json:"processing_started_at" db:"processing_started_at"`
	SequenceStepID       *string    `json:"sequence_stepid" db:"sequence_stepid"`
	UserID               *string    `json:"user_id" db:"user_id"`
}

// SequenceLog model for tracking message history
type SequenceLog struct {
	ID           string    `json:"id" db:"id"`
	SequenceID   string    `json:"sequence_id" db:"sequence_id"`
	ContactID    string    `json:"contact_id" db:"contact_id"`
	StepID       string    `json:"step_id" db:"step_id"`
	Day          int       `json:"day" db:"day"`
	Status       string    `json:"status" db:"status"` // sent, failed
	MessageID    string    `json:"message_id" db:"message_id"`
	ErrorMessage string    `json:"error_message" db:"error_message"`
	SentAt       time.Time `json:"sent_at" db:"sent_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}
