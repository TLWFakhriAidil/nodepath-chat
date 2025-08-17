package models

import (
	"time"

	"github.com/google/uuid"
)

// TeamMember represents a team member who can login and view assigned devices
type TeamMember struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`   // matches device name
	Password  string    `json:"password" db:"password"`   // plain text
	CreatedBy uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	IsActive  bool      `json:"is_active" db:"is_active"`
}

// TeamSession represents a login session for team members
type TeamSession struct {
	ID           uuid.UUID `json:"id" db:"id"`
	TeamMemberID uuid.UUID `json:"team_member_id" db:"team_member_id"`
	Token        string    `json:"token" db:"token"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// TeamMemberWithDevices includes the devices that the team member can access
type TeamMemberWithDevices struct {
	TeamMember
	DeviceCount int      `json:"device_count"`
	DeviceIDs   []string `json:"device_ids"`
}
