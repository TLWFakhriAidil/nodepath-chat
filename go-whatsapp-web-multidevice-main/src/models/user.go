package models

import (
	"time"
)

// User represents a system user
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	FullName     string    `json:"fullName"`
	PasswordHash string    `json:"-"` // Don't expose password hash in JSON
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	IsActive     bool      `json:"isActive"`
	LastLogin    time.Time `json:"lastLogin"`
}

// UserDevice represents a WhatsApp device belonging to a user
type UserDevice struct {
	ID              string    `json:"id"`
	UserID          string    `json:"userId"`
	DeviceName      string    `json:"deviceName"`
	Phone           string    `json:"phone"`
	Status          string    `json:"status"` // online, offline, connecting
	LastSeen        time.Time `json:"lastSeen"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	JID             string    `json:"jid,omitempty"`
	MinDelaySeconds int       `json:"minDelaySeconds"`
	MaxDelaySeconds int       `json:"maxDelaySeconds"`
	Platform        string    `json:"platform"` // Whacenter, etc.
}

// UserSession represents an active user session
type UserSession struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}