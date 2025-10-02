package repository

import (
	"database/sql"
	"fmt"
	"time"

	"nodepath-chat/internal/models"

	"github.com/sirupsen/logrus"
)

// DeviceSettingsRepository interface defines methods for device settings management
type DeviceSettingsRepository interface {
	// Create operations
	CreateDeviceSettings(settings *models.DeviceSettings) error

	// Read operations
	GetDeviceSettingsByID(deviceID string) (*models.DeviceSettings, error)
	GetDeviceSettingsByDevice(idDevice string) (*models.DeviceSettings, error)
	GetAllDeviceSettings() ([]models.DeviceSettings, error)
	GetDeviceSettingsByProvider(provider string) ([]models.DeviceSettings, error)
	GetAPIKeyByDevice(idDevice string) (string, error)
	GetProviderByDevice(idDevice string) (string, error)
	GetAPIKeyOptionByDevice(idDevice string) (string, error)

	// Update operations
	UpdateDeviceSettings(settings *models.DeviceSettings) error
	UpdateAPIKey(deviceID string, apiKey string) error
	UpdateProvider(deviceID string, provider string) error
	UpdateAPIKeyOption(deviceID string, apiKeyOption string) error
	UpdateWebhookID(deviceID string, webhookID string) error

	// Delete operations
	DeleteDeviceSettings(deviceID string) error

	// Utility operations
	DeviceExists(idDevice string) (bool, error)
	GetDeviceCount() (int, error)
}

// deviceSettingsRepository implements DeviceSettingsRepository interface
type deviceSettingsRepository struct {
	db *sql.DB
}

// NewDeviceSettingsRepository creates a new instance of DeviceSettingsRepository
func NewDeviceSettingsRepository(db *sql.DB) DeviceSettingsRepository {
	return &deviceSettingsRepository{
		db: db,
	}
}

// CreateDeviceSettings creates a new device settings record
func (r *deviceSettingsRepository) CreateDeviceSettings(settings *models.DeviceSettings) error {
	settings.CreatedAt = time.Now()
	settings.UpdatedAt = time.Now()

	query := `
		INSERT INTO device_setting_nodepath (
			device_id, api_key_option, webhook_id, provider, 
			api_key, id_device, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		settings.DeviceID, settings.APIKeyOption, settings.WebhookID, settings.Provider,
		settings.APIKey, settings.IDDevice, settings.CreatedAt, settings.UpdatedAt,
	)

	if err != nil {
		logrus.WithError(err).Error("Failed to create device settings")
		return fmt.Errorf("failed to create device settings: %w", err)
	}

	logrus.WithField("device_id", settings.DeviceID).Info("Device settings created successfully")
	return nil
}

// GetDeviceSettingsByID retrieves device settings by device_id
func (r *deviceSettingsRepository) GetDeviceSettingsByID(deviceID string) (*models.DeviceSettings, error) {
	// Check if database connection is available
	if r.db == nil {
		return nil, fmt.Errorf("database connection is not available")
	}

	query := `
		SELECT device_id, api_key_option, webhook_id, provider, 
		       api_key, id_device, created_at, updated_at
		FROM device_setting_nodepath 
		WHERE device_id = ?
	`

	row := r.db.QueryRow(query, deviceID)

	settings := &models.DeviceSettings{}
	err := row.Scan(
		&settings.DeviceID, &settings.APIKeyOption, &settings.WebhookID, &settings.Provider,
		&settings.APIKey, &settings.IDDevice, &settings.CreatedAt, &settings.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		logrus.WithError(err).Error("Failed to get device settings by ID")
		return nil, fmt.Errorf("failed to get device settings: %w", err)
	}

	return settings, nil
}

// GetDeviceSettingsByDevice retrieves device settings by id_device
func (r *deviceSettingsRepository) GetDeviceSettingsByDevice(idDevice string) (*models.DeviceSettings, error) {
	// Check if database connection is available
	if r.db == nil {
		return nil, fmt.Errorf("database connection is not available")
	}

	query := `
		SELECT device_id, api_key_option, webhook_id, provider, 
		       api_key, id_device, created_at, updated_at
		FROM device_setting_nodepath 
		WHERE id_device = ?
	`

	row := r.db.QueryRow(query, idDevice)

	settings := &models.DeviceSettings{}
	err := row.Scan(
		&settings.DeviceID, &settings.APIKeyOption, &settings.WebhookID, &settings.Provider,
		&settings.APIKey, &settings.IDDevice, &settings.CreatedAt, &settings.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		logrus.WithError(err).Error("Failed to get device settings by device")
		return nil, fmt.Errorf("failed to get device settings: %w", err)
	}

	return settings, nil
}

// GetAllDeviceSettings retrieves all device settings
func (r *deviceSettingsRepository) GetAllDeviceSettings() ([]models.DeviceSettings, error) {
	// Check if database connection is available
	if r.db == nil {
		return nil, fmt.Errorf("database connection is not available")
	}

	query := `
		SELECT device_id, api_key_option, webhook_id, provider, 
		       api_key, id_device, created_at, updated_at
		FROM device_setting_nodepath 
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		logrus.WithError(err).Error("Failed to get all device settings")
		return nil, fmt.Errorf("failed to get all device settings: %w", err)
	}
	defer rows.Close()

	var settingsList []models.DeviceSettings
	for rows.Next() {
		settings := models.DeviceSettings{}
		err := rows.Scan(
			&settings.DeviceID, &settings.APIKeyOption, &settings.WebhookID, &settings.Provider,
			&settings.APIKey, &settings.IDDevice, &settings.CreatedAt, &settings.UpdatedAt,
		)

		if err != nil {
			logrus.WithError(err).Error("Failed to scan device settings")
			continue
		}

		settingsList = append(settingsList, settings)
	}

	return settingsList, nil
}

// GetDeviceSettingsByProvider retrieves device settings by provider
func (r *deviceSettingsRepository) GetDeviceSettingsByProvider(provider string) ([]models.DeviceSettings, error) {
	query := `
		SELECT device_id, api_key_option, webhook_id, provider, 
		       api_key, id_device, created_at, updated_at
		FROM device_setting_nodepath 
		WHERE provider = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, provider)
	if err != nil {
		logrus.WithError(err).Error("Failed to get device settings by provider")
		return nil, fmt.Errorf("failed to get device settings by provider: %w", err)
	}
	defer rows.Close()

	var settingsList []models.DeviceSettings
	for rows.Next() {
		settings := models.DeviceSettings{}
		err := rows.Scan(
			&settings.DeviceID, &settings.APIKeyOption, &settings.WebhookID, &settings.Provider,
			&settings.APIKey, &settings.IDDevice, &settings.CreatedAt, &settings.UpdatedAt,
		)

		if err != nil {
			logrus.WithError(err).Error("Failed to scan device settings")
			continue
		}

		settingsList = append(settingsList, settings)
	}

	return settingsList, nil
}

// GetAPIKeyByDevice retrieves API key for a specific device
func (r *deviceSettingsRepository) GetAPIKeyByDevice(idDevice string) (string, error) {
	query := `SELECT api_key FROM device_setting_nodepath WHERE id_device = ?`

	var apiKey string
	row := r.db.QueryRow(query, idDevice)
	err := row.Scan(&apiKey)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // Not found
		}
		logrus.WithError(err).Error("Failed to get API key by device")
		return "", fmt.Errorf("failed to get API key: %w", err)
	}

	return apiKey, nil
}

// GetProviderByDevice retrieves provider for a specific device
func (r *deviceSettingsRepository) GetProviderByDevice(idDevice string) (string, error) {
	query := `SELECT provider FROM device_setting_nodepath WHERE id_device = ?`

	var provider string
	row := r.db.QueryRow(query, idDevice)
	err := row.Scan(&provider)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // Not found
		}
		logrus.WithError(err).Error("Failed to get provider by device")
		return "", fmt.Errorf("failed to get provider: %w", err)
	}

	return provider, nil
}

// GetAPIKeyOptionByDevice retrieves API key option for a specific device
func (r *deviceSettingsRepository) GetAPIKeyOptionByDevice(idDevice string) (string, error) {
	query := `SELECT api_key_option FROM device_setting_nodepath WHERE id_device = ?`

	var apiKeyOption string
	row := r.db.QueryRow(query, idDevice)
	err := row.Scan(&apiKeyOption)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // Not found
		}
		logrus.WithError(err).Error("Failed to get API key option by device")
		return "", fmt.Errorf("failed to get API key option: %w", err)
	}

	return apiKeyOption, nil
}

// UpdateDeviceSettings updates an existing device settings record
func (r *deviceSettingsRepository) UpdateDeviceSettings(settings *models.DeviceSettings) error {
	settings.UpdatedAt = time.Now()

	query := `
		UPDATE device_setting_nodepath SET 
			api_key_option = ?, webhook_id = ?, provider = ?, 
			api_key = ?, id_device = ?, updated_at = ?
		WHERE device_id = ?
	`

	_, err := r.db.Exec(query,
		settings.APIKeyOption, settings.WebhookID, settings.Provider,
		settings.APIKey, settings.IDDevice, settings.UpdatedAt, settings.DeviceID,
	)

	if err != nil {
		logrus.WithError(err).Error("Failed to update device settings")
		return fmt.Errorf("failed to update device settings: %w", err)
	}

	logrus.WithField("device_id", settings.DeviceID).Info("Device settings updated successfully")
	return nil
}

// UpdateAPIKey updates the API key for a specific device
func (r *deviceSettingsRepository) UpdateAPIKey(deviceID string, apiKey string) error {
	query := `
		UPDATE device_setting_nodepath 
		SET api_key = ?, updated_at = ?
		WHERE device_id = ?
	`

	_, err := r.db.Exec(query, apiKey, time.Now(), deviceID)
	if err != nil {
		logrus.WithError(err).Error("Failed to update API key")
		return fmt.Errorf("failed to update API key: %w", err)
	}

	logrus.WithField("device_id", deviceID).Info("API key updated successfully")
	return nil
}

// UpdateProvider updates the provider for a specific device
func (r *deviceSettingsRepository) UpdateProvider(deviceID string, provider string) error {
	query := `
		UPDATE device_setting_nodepath 
		SET provider = ?, updated_at = ?
		WHERE device_id = ?
	`

	_, err := r.db.Exec(query, provider, time.Now(), deviceID)
	if err != nil {
		logrus.WithError(err).Error("Failed to update provider")
		return fmt.Errorf("failed to update provider: %w", err)
	}

	logrus.WithField("device_id", deviceID).Info("Provider updated successfully")
	return nil
}

// UpdateAPIKeyOption updates the API key option for a specific device
func (r *deviceSettingsRepository) UpdateAPIKeyOption(deviceID string, apiKeyOption string) error {
	query := `
		UPDATE device_setting_nodepath 
		SET api_key_option = ?, updated_at = ?
		WHERE device_id = ?
	`

	_, err := r.db.Exec(query, apiKeyOption, time.Now(), deviceID)
	if err != nil {
		logrus.WithError(err).Error("Failed to update API key option")
		return fmt.Errorf("failed to update API key option: %w", err)
	}

	logrus.WithField("device_id", deviceID).Info("API key option updated successfully")
	return nil
}

// UpdateWebhookID updates the webhook ID for a specific device
func (r *deviceSettingsRepository) UpdateWebhookID(deviceID string, webhookID string) error {
	query := `
		UPDATE device_setting_nodepath 
		SET webhook_id = ?, updated_at = ?
		WHERE device_id = ?
	`

	_, err := r.db.Exec(query, webhookID, time.Now(), deviceID)
	if err != nil {
		logrus.WithError(err).Error("Failed to update webhook ID")
		return fmt.Errorf("failed to update webhook ID: %w", err)
	}

	logrus.WithField("device_id", deviceID).Info("Webhook ID updated successfully")
	return nil
}

// DeleteDeviceSettings deletes device settings by device_id
func (r *deviceSettingsRepository) DeleteDeviceSettings(deviceID string) error {
	query := `DELETE FROM device_setting_nodepath WHERE device_id = ?`

	_, err := r.db.Exec(query, deviceID)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete device settings")
		return fmt.Errorf("failed to delete device settings: %w", err)
	}

	logrus.WithField("device_id", deviceID).Info("Device settings deleted successfully")
	return nil
}

// DeviceExists checks if a device exists in the settings
func (r *deviceSettingsRepository) DeviceExists(idDevice string) (bool, error) {
	query := `SELECT COUNT(*) FROM device_setting_nodepath WHERE id_device = ?`

	var count int
	row := r.db.QueryRow(query, idDevice)
	err := row.Scan(&count)
	if err != nil {
		logrus.WithError(err).Error("Failed to check if device exists")
		return false, fmt.Errorf("failed to check device existence: %w", err)
	}

	return count > 0, nil
}

// GetDeviceCount returns the total number of configured devices
func (r *deviceSettingsRepository) GetDeviceCount() (int, error) {
	query := `SELECT COUNT(*) FROM device_setting_nodepath`

	var count int
	row := r.db.QueryRow(query)
	err := row.Scan(&count)
	if err != nil {
		logrus.WithError(err).Error("Failed to get device count")
		return 0, fmt.Errorf("failed to get device count: %w", err)
	}

	return count, nil
}
