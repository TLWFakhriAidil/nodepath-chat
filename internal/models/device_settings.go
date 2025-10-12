package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// DeviceSettings represents a device configuration
type DeviceSettings struct {
	ID           string         `json:"id" db:"id"`
	DeviceID     sql.NullString `json:"-" db:"device_id"`
	APIKeyOption string         `json:"api_key_option" db:"api_key_option"`
	WebhookID    sql.NullString `json:"-" db:"webhook_id"`
	Provider     string         `json:"provider" db:"provider"`
	PhoneNumber  sql.NullString `json:"-" db:"phone_number"`
	APIKey       sql.NullString `json:"-" db:"api_key"`
	IDDevice     sql.NullString `json:"-" db:"id_device"`
	IDERP        sql.NullString `json:"-" db:"id_erp"`
	IDAdmin      sql.NullString `json:"-" db:"id_admin"`
	UserID       sql.NullString `json:"-" db:"user_id"`
	Instance     sql.NullString `json:"-" db:"instance"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
}

// MarshalJSON implements custom JSON marshaling for DeviceSettings
func (d *DeviceSettings) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"id":             d.ID,
		"device_id":      nullStringToString(d.DeviceID),
		"api_key_option": d.APIKeyOption,
		"webhook_id":     nullStringToString(d.WebhookID),
		"provider":       d.Provider,
		"phone_number":   nullStringToString(d.PhoneNumber),
		"api_key":        nullStringToString(d.APIKey),
		"id_device":      nullStringToString(d.IDDevice),
		"id_erp":         nullStringToString(d.IDERP),
		"id_admin":       nullStringToString(d.IDAdmin),
		"user_id":        nullStringToString(d.UserID),
		"instance":       nullStringToString(d.Instance),
		"created_at":     d.CreatedAt,
		"updated_at":     d.UpdatedAt,
	})
}

// nullStringToString converts sql.NullString to string
func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// nullInt32ToInt converts sql.NullInt32 to *int
func nullInt32ToInt(ni sql.NullInt32) *int {
	if ni.Valid {
		val := int(ni.Int32)
		return &val
	}
	return nil
}

// CreateDeviceSettingsRequest represents the request to create device settings
type CreateDeviceSettingsRequest struct {
	DeviceID     string `json:"device_id"` // Optional - can be empty for manual creation
	APIKeyOption string `json:"api_key_option"`
	WebhookID    string `json:"webhook_id"`
	Provider     string `json:"provider"`
	PhoneNumber  string `json:"phone_number"`
	APIKey       string `json:"api_key"`
	IDDevice     string `json:"id_device" validate:"required"`
	IDERP        string `json:"id_erp" validate:"required"`
	IDAdmin      string `json:"id_admin" validate:"required"`
	UserID       string `json:"user_id"`
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
	UserID       string `json:"user_id"`
	Instance     string `json:"instance"`
}
