package sequence

import (
	"time"
)

// ISequenceUsecase interface for sequence operations
type ISequenceUsecase interface {
	CreateSequence(request CreateSequenceRequest) (SequenceResponse, error)
	GetSequences(userID string) ([]SequenceResponse, error)
	GetSequenceByID(sequenceID string) (SequenceDetailResponse, error)
	UpdateSequence(sequenceID string, request UpdateSequenceRequest) error
	DeleteSequence(sequenceID string) error
	
	// Contact management
	AddContactsToSequence(sequenceID string, contacts []string) error
	RemoveContactFromSequence(sequenceID string, contactID string) error
	GetSequenceContacts(sequenceID string) ([]SequenceContactResponse, error)
	
	// Execution
	StartSequence(sequenceID string) error
	PauseSequence(sequenceID string) error
	ProcessSequences() error // Called by cron job
}

// CreateSequenceRequest for creating new sequence - simplified
type CreateSequenceRequest struct {
	Name            string                      `json:"name" validate:"required"`
	Description     string                      `json:"description"`
	UserID          string                      `json:"user_id"`
	DeviceID        *string                     `json:"device_id"` // Optional - sequences use all user devices
	Niche           string                      `json:"niche"`
	Status          string                      `json:"status"`
	Trigger         string                      `json:"trigger"`     // Main trigger for the sequence
	StartTrigger    string                      `json:"start_trigger"` // Deprecated
	EndTrigger      string                      `json:"end_trigger"`   // Deprecated
	IsActive        bool                        `json:"is_active"`
	TimeSchedule    string                      `json:"time_schedule"`
	MinDelaySeconds int                         `json:"min_delay_seconds"`
	MaxDelaySeconds int                         `json:"max_delay_seconds"`
	Steps           []CreateSequenceStepRequest `json:"steps" validate:"required,min=1"`
}

// CreateSequenceStepRequest for each step
type CreateSequenceStepRequest struct {
	Day               int    `json:"day"`
	DayNumber         int    `json:"day_number" validate:"required,min=1"`
	Trigger           string `json:"trigger"`
	NextTrigger       string `json:"next_trigger"`
	TriggerDelayHours int    `json:"trigger_delay_hours"`
	IsEntryPoint      bool   `json:"is_entry_point"`
	MessageType       string `json:"message_type"`
	SendTime          string `json:"send_time"`
	TimeSchedule      string `json:"time_schedule"`
	Content           string `json:"content"`
	ImageURL          string `json:"image_url"`
	MediaURL          string `json:"media_url"`
	Caption           string `json:"caption"`
	MinDelaySeconds   int    `json:"min_delay_seconds"`
	MaxDelaySeconds   int    `json:"max_delay_seconds"`
}

// UpdateSequenceRequest for updating sequence
type UpdateSequenceRequest struct {
	Name            string                      `json:"name"`
	Description     string                      `json:"description"`
	Niche           string                      `json:"niche"`
	Status          string                      `json:"status"`
	Trigger         string                      `json:"trigger"`     // Main trigger for the sequence
	StartTrigger    string                      `json:"start_trigger"` // Deprecated
	EndTrigger      string                      `json:"end_trigger"`   // Deprecated
	IsActive        bool                        `json:"is_active"`
	TimeSchedule    string                      `json:"time_schedule"`
	MinDelaySeconds int                         `json:"min_delay_seconds"`
	MaxDelaySeconds int                         `json:"max_delay_seconds"`
	Steps           []CreateSequenceStepRequest `json:"steps"`
}

// SequenceResponse basic sequence info
type SequenceResponse struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	UserID          string                 `json:"user_id"`
	DeviceID        *string                `json:"device_id"` // Optional - sequences use all user devices
	Niche           string                 `json:"niche"`
	Status          string                 `json:"status"`
	Trigger         string                 `json:"trigger"`      // Main trigger for the sequence
	StartTrigger    string                 `json:"start_trigger"` // Deprecated
	EndTrigger      string                 `json:"end_trigger"`   // Deprecated
	TotalSteps      int                    `json:"total_steps"`
	TotalDays       int                    `json:"total_days"`
	IsActive        bool                   `json:"is_active"`
	TimeSchedule    string                 `json:"time_schedule"`
	MinDelaySeconds int                    `json:"min_delay_seconds"`
	MaxDelaySeconds int                    `json:"max_delay_seconds"`
	ContactCount    int                    `json:"contact_count"`
	ContactsCount   int                    `json:"contacts_count"`
	StepCount       int                    `json:"step_count"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	Steps           []SequenceStepResponse `json:"steps"`
}

// SequenceStepResponse for each step
type SequenceStepResponse struct {
	ID                string `json:"id"`
	SequenceID        string `json:"sequence_id"`
	DayNumber         int    `json:"day_number"`
	Trigger           string `json:"trigger"`
	NextTrigger       string `json:"next_trigger"`
	TriggerDelayHours int    `json:"trigger_delay_hours"`
	IsEntryPoint      bool   `json:"is_entry_point"`
	MessageType       string `json:"message_type"`
	TimeSchedule      string `json:"time_schedule"`
	Content           string `json:"content"`
	MediaURL          string `json:"media_url"`
	Caption           string `json:"caption"`
	MinDelaySeconds   int    `json:"min_delay_seconds"`
	MaxDelaySeconds   int    `json:"max_delay_seconds"`
}

// SequenceStats statistics for a sequence
type SequenceStats struct {
	TotalContacts    int `json:"total_contacts"`
	ActiveContacts   int `json:"active_contacts"`
	CompletedContacts int `json:"completed_contacts"`
	PausedContacts   int `json:"paused_contacts"`
	TotalMessagesSent int `json:"total_messages_sent"`
	MessagesSent     int `json:"messages_sent"`
	SuccessRate      float64 `json:"success_rate"`
}

// SequenceDetailResponse includes full details
type SequenceDetailResponse struct {
	SequenceResponse
	Contacts []SequenceContactResponse `json:"contacts"`
	Stats    SequenceStats            `json:"stats"`
}

// SequenceContactResponse contact info
type SequenceContactResponse struct {
	ID           string     `json:"id"`
	ContactPhone string     `json:"contact_phone"`
	ContactName  string     `json:"contact_name"`
	CurrentStep  int        `json:"current_step"`
	Status       string     `json:"status"`
	AddedAt      *time.Time `json:"added_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

// AddContactsRequest for adding contacts
type AddContactsRequest struct {
	Contacts []string `json:"contacts" validate:"required,min=1"`
}
