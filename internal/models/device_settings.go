package models

import (
	"time"
)

// DeviceSettings represents a device configuration
type DeviceSettings struct {
	ID           string    `json:"id" db:"id"`
	DeviceID     string    `json:"device_id" db:"device_id"`
	APIKeyOption string    `json:"api_key_option" db:"api_key_option"`
	WebhookID    string    `json:"webhook_id" db:"webhook_id"`
	Provider     string    `json:"provider" db:"provider"`
	PhoneNumber  string    `json:"phone_number" db:"phone_number"`
	APIKey       string    `json:"api_key" db:"api_key"`
	IDDevice     string    `json:"id_device" db:"id_device"`
	IDERP        string    `json:"id_erp" db:"id_erp"`
	IDAdmin      string    `json:"id_admin" db:"id_admin"`
	Instance     string    `json:"instance" db:"instance"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// CreateDeviceSettingsRequest represents the request to create device settings
type CreateDeviceSettingsRequest struct {
	DeviceID     string `json:"device_id" validate:"required"`
	APIKeyOption string `json:"api_key_option"`
	WebhookID    string `json:"webhook_id"`
	Provider     string `json:"provider"`
	PhoneNumber  string `json:"phone_number"`
	APIKey       string `json:"api_key"`
	IDDevice     string `json:"id_device" validate:"required"`
	IDERP        string `json:"id_erp" validate:"required"`
	IDAdmin      string `json:"id_admin" validate:"required"`
	Instance     string `json:"instance"`
}

// UpdateDeviceSettingsRequest represents the request to update device settings
type UpdateDeviceSettingsRequest struct {
	DeviceID     string `json:"device_id"`
	APIKeyOption string `json:"api_key_option"`
	WebhookID    string `json:"webhook_id"`
	Provider     string `json:"provider"`
	PhoneNumber  string `json:"phone_number"`
	APIKey       string `json:"api_key"`
	IDDevice     string `json:"id_device"`
	IDERP        string `json:"id_erp"`
	IDAdmin      string `json:"id_admin"`
	Instance     string `json:"instance"`
}