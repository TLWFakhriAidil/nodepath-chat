package services

import (
	"database/sql"
	"fmt"
	"time"

	"nodepath-chat/internal/models"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// DeviceSettingsService handles device settings operations
type DeviceSettingsService struct {
	db *sql.DB
}

// NewDeviceSettingsService creates a new device settings service
func NewDeviceSettingsService(db *sql.DB) *DeviceSettingsService {
	return &DeviceSettingsService{
		db: db,
	}
}

// GetAll retrieves all device settings
func (s *DeviceSettingsService) GetAll() ([]*models.DeviceSettings, error) {
	query := `
		SELECT id, device_id, api_key_option, webhook_id, provider, phone_number, api_key, 
		       id_device, id_erp, id_admin, instance, created_at, updated_at
		FROM device_setting_nodepath
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query device settings: %w", err)
	}
	defer rows.Close()

	var settings []*models.DeviceSettings
	for rows.Next() {
		setting := &models.DeviceSettings{}
		err := rows.Scan(
			&setting.ID,
			&setting.DeviceID,
			&setting.APIKeyOption,
			&setting.WebhookID,
			&setting.Provider,
			&setting.PhoneNumber,
			&setting.APIKey,
			&setting.IDDevice,
			&setting.IDERP,
			&setting.IDAdmin,
			&setting.Instance,
			&setting.CreatedAt,
			&setting.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device setting: %w", err)
		}
		settings = append(settings, setting)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating device settings: %w", err)
	}

	return settings, nil
}

// GetByID retrieves a device setting by ID
func (s *DeviceSettingsService) GetByID(id string) (*models.DeviceSettings, error) {
	query := `
		SELECT id, device_id, api_key_option, webhook_id, provider, phone_number, api_key, 
		       id_device, id_erp, id_admin, instance, created_at, updated_at
		FROM device_setting_nodepath
		WHERE id = ?
	`

	setting := &models.DeviceSettings{}
	err := s.db.QueryRow(query, id).Scan(
		&setting.ID,
		&setting.DeviceID,
		&setting.APIKeyOption,
		&setting.WebhookID,
		&setting.Provider,
		&setting.PhoneNumber,
		&setting.APIKey,
		&setting.IDDevice,
		&setting.IDERP,
		&setting.IDAdmin,
		&setting.Instance,
		&setting.CreatedAt,
		&setting.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device setting not found")
		}
		return nil, fmt.Errorf("failed to get device setting: %w", err)
	}

	return setting, nil
}

// Create creates a new device setting
func (s *DeviceSettingsService) Create(req *models.CreateDeviceSettingsRequest) (*models.DeviceSettings, error) {
	id := uuid.New().String()
	now := time.Now()

	// Set defaults if not provided
	apiKeyOption := req.APIKeyOption
	if apiKeyOption == "" {
		apiKeyOption = "chat_gpt_4_1_new"
	}

	provider := req.Provider
	if provider == "" {
		provider = "wablas"
	}

	query := `
		INSERT INTO device_setting_nodepath 
		(id, device_id, api_key_option, webhook_id, provider, phone_number, api_key, id_device, id_erp, id_admin, instance, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		id,
		req.DeviceID,
		apiKeyOption,
		req.WebhookID,
		provider,
		req.PhoneNumber,
		req.APIKey,
		req.IDDevice,
		req.IDERP,
		req.IDAdmin,
		req.Instance,
		now,
		now,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create device setting: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"id":         id,
		"device_id":  req.DeviceID,
		"id_device":  req.IDDevice,
		"id_erp":     req.IDERP,
		"id_admin":   req.IDAdmin,
	}).Info("Device setting created")

	return s.GetByID(id)
}

// Update updates an existing device setting
func (s *DeviceSettingsService) Update(id string, req *models.UpdateDeviceSettingsRequest) (*models.DeviceSettings, error) {
	// Check if device setting exists
	existing, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.DeviceID != "" {
		existing.DeviceID = req.DeviceID
	}
	if req.APIKeyOption != "" {
		existing.APIKeyOption = req.APIKeyOption
	}
	if req.WebhookID != "" {
		existing.WebhookID = req.WebhookID
	}
	if req.Provider != "" {
		existing.Provider = req.Provider
	}
	if req.PhoneNumber != "" {
		existing.PhoneNumber = req.PhoneNumber
	}
	if req.APIKey != "" {
		existing.APIKey = req.APIKey
	}
	if req.IDDevice != "" {
		existing.IDDevice = req.IDDevice
	}
	if req.IDERP != "" {
		existing.IDERP = req.IDERP
	}
	if req.IDAdmin != "" {
		existing.IDAdmin = req.IDAdmin
	}
	if req.Instance != "" {
		existing.Instance = req.Instance
	}

	existing.UpdatedAt = time.Now()

	query := `
		UPDATE device_setting_nodepath 
		SET device_id = ?, api_key_option = ?, webhook_id = ?, provider = ?, phone_number = ?, api_key = ?, 
		    id_device = ?, id_erp = ?, id_admin = ?, instance = ?, updated_at = ?
		WHERE id = ?
	`

	_, err = s.db.Exec(query,
		existing.DeviceID,
		existing.APIKeyOption,
		existing.WebhookID,
		existing.Provider,
		existing.PhoneNumber,
		existing.APIKey,
		existing.IDDevice,
		existing.IDERP,
		existing.IDAdmin,
		existing.Instance,
		existing.UpdatedAt,
		id,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update device setting: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"id":         id,
		"device_id":  existing.DeviceID,
		"id_device":  existing.IDDevice,
		"id_erp":     existing.IDERP,
		"id_admin":   existing.IDAdmin,
	}).Info("Device setting updated")

	return existing, nil
}

// Delete deletes a device setting
func (s *DeviceSettingsService) Delete(id string) error {
	// Check if device setting exists
	_, err := s.GetByID(id)
	if err != nil {
		return err
	}

	query := `DELETE FROM device_setting_nodepath WHERE id = ?`
	_, err = s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete device setting: %w", err)
	}

	logrus.WithField("id", id).Info("Device setting deleted")
	return nil
}