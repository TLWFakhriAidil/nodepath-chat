package services

import (
	"database/sql"
	"fmt"
	"time"

	"nodepath-chat/internal/models"
	"nodepath-chat/internal/utils"

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
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	query := `
		SELECT id, device_id, api_key_option, webhook_id, provider, phone_number, api_key, 
		       id_device, id_erp, id_admin, instance, created_at, updated_at, user_id
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
			&setting.UserID,
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

// GetByUserID retrieves device settings for a specific user (deprecated - use GetByUserIDString)
func (s *DeviceSettingsService) GetByUserID(userID int) ([]*models.DeviceSettings, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	query := `
		SELECT id, device_id, api_key_option, webhook_id, provider, phone_number, api_key, 
		       id_device, id_erp, id_admin, instance, created_at, updated_at, user_id
		FROM device_setting_nodepath
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query device settings for user: %w", err)
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
			&setting.UserID,
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

// GetByUserIDString retrieves device settings for a specific user by UUID string
func (s *DeviceSettingsService) GetByUserIDString(userID string) ([]*models.DeviceSettings, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	query := `
		SELECT id, device_id, api_key_option, webhook_id, provider, phone_number, api_key, 
		       id_device, id_erp, id_admin, instance, created_at, updated_at, user_id
		FROM device_setting_nodepath
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query device settings for user: %w", err)
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
			&setting.UserID,
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
	// Check if database connection is available
	if s.db == nil {
		return nil, fmt.Errorf("database connection is not available")
	}

	query := `
		SELECT id, device_id, api_key_option, webhook_id, provider, phone_number, api_key, 
		       id_device, id_erp, id_admin, instance, created_at, updated_at, user_id
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
		&setting.UserID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device setting not found")
		}
		return nil, fmt.Errorf("failed to get device setting: %w", err)
	}

	return setting, nil
}

// GetByIDDevice retrieves a device setting by id_device field
func (s *DeviceSettingsService) GetByIDDevice(idDevice string) (*models.DeviceSettings, error) {
	// Check if database connection is available
	if s.db == nil {
		return nil, fmt.Errorf("database connection is not available")
	}

	query := `
		SELECT id, device_id, api_key_option, webhook_id, provider, phone_number, api_key, 
		       id_device, id_erp, id_admin, instance, created_at, updated_at, user_id
		FROM device_setting_nodepath
		WHERE id_device = ?
		ORDER BY created_at DESC
		LIMIT 1
	`

	setting := &models.DeviceSettings{}
	err := s.db.QueryRow(query, idDevice).Scan(
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
		&setting.UserID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device setting not found")
		}
		return nil, fmt.Errorf("failed to get device setting: %w", err)
	}

	return setting, nil
}

// Upsert creates a new device setting or updates existing one based on id_device
// Uses database transactions to ensure data consistency during upsert operations
func (s *DeviceSettingsService) Upsert(req *models.CreateDeviceSettingsRequest) (*models.DeviceSettings, error) {
	// Check if database connection is available
	if s.db == nil {
		return nil, fmt.Errorf("database connection is not available")
	}

	var resultID string

	err := utils.WithTransaction(s.db, func(tx *sql.Tx) error {
		// Check if a device setting already exists for this id_device within transaction
		var existingID string
		checkQuery := `
			SELECT id 
			FROM device_setting_nodepath 
			WHERE id_device = ?
			FOR UPDATE
		`
		err := tx.QueryRow(checkQuery, req.IDDevice).Scan(&existingID)

		if err == nil {
			// Device setting exists, update it within transaction
			now := time.Now()

			// Set defaults if not provided
			apiKeyOption := req.APIKeyOption
			if apiKeyOption == "" {
				apiKeyOption = "openai/gpt-4.1"
			}

			provider := req.Provider
			if provider == "" {
				provider = "wablas"
			}

			// Convert strings to sql.NullString for nullable fields
			var deviceID, webhookID, phoneNumber, apiKey, idDevice, idERP, idAdmin, instance, userID sql.NullString

			if req.DeviceID != "" {
				deviceID = sql.NullString{String: req.DeviceID, Valid: true}
			}
			if req.WebhookID != "" {
				webhookID = sql.NullString{String: req.WebhookID, Valid: true}
			}
			if req.PhoneNumber != "" {
				phoneNumber = sql.NullString{String: req.PhoneNumber, Valid: true}
			}
			if req.APIKey != "" {
				apiKey = sql.NullString{String: req.APIKey, Valid: true}
			}
			if req.IDDevice != "" {
				idDevice = sql.NullString{String: req.IDDevice, Valid: true}
			}
			if req.IDERP != "" {
				idERP = sql.NullString{String: req.IDERP, Valid: true}
			}
			if req.IDAdmin != "" {
				idAdmin = sql.NullString{String: req.IDAdmin, Valid: true}
			}
			if req.Instance != "" {
				instance = sql.NullString{String: req.Instance, Valid: true}
			}
			if req.UserID != "" {
				userID = sql.NullString{String: req.UserID, Valid: true}
			}

			updateQuery := `
				UPDATE device_setting_nodepath 
				SET device_id = ?, api_key_option = ?, webhook_id = ?, provider = ?, phone_number = ?, api_key = ?, 
				    id_device = ?, id_erp = ?, id_admin = ?, instance = ?, updated_at = ?, user_id = ?
				WHERE id = ?
			`

			_, err = tx.Exec(updateQuery,
				deviceID, apiKeyOption, webhookID, provider, phoneNumber, apiKey,
				idDevice, idERP, idAdmin, instance, now, userID, existingID,
			)

			if err != nil {
				return fmt.Errorf("failed to update device setting: %w", err)
			}

			resultID = existingID
			logrus.WithFields(logrus.Fields{
				"id":        existingID,
				"device_id": req.DeviceID,
				"id_device": req.IDDevice,
				"id_erp":    req.IDERP,
				"id_admin":  req.IDAdmin,
			}).Info("Device setting updated")

		} else if err == sql.ErrNoRows {
			// Device setting doesn't exist, create new one within transaction
			id := uuid.New().String()
			now := time.Now()

			// Set defaults if not provided
			apiKeyOption := req.APIKeyOption
			if apiKeyOption == "" {
				apiKeyOption = "openai/gpt-4.1"
			}

			provider := req.Provider
			if provider == "" {
				provider = "wablas"
			}

			// Convert strings to sql.NullString for nullable fields
			var deviceID, webhookID, phoneNumber, apiKey, idDevice, idERP, idAdmin, instance, userID sql.NullString

			if req.DeviceID != "" {
				deviceID = sql.NullString{String: req.DeviceID, Valid: true}
			}
			if req.WebhookID != "" {
				webhookID = sql.NullString{String: req.WebhookID, Valid: true}
			}
			if req.PhoneNumber != "" {
				phoneNumber = sql.NullString{String: req.PhoneNumber, Valid: true}
			}
			if req.APIKey != "" {
				apiKey = sql.NullString{String: req.APIKey, Valid: true}
			}
			if req.IDDevice != "" {
				idDevice = sql.NullString{String: req.IDDevice, Valid: true}
			}
			if req.IDERP != "" {
				idERP = sql.NullString{String: req.IDERP, Valid: true}
			}
			if req.IDAdmin != "" {
				idAdmin = sql.NullString{String: req.IDAdmin, Valid: true}
			}
			if req.Instance != "" {
				instance = sql.NullString{String: req.Instance, Valid: true}
			}
			if req.UserID != "" {
				userID = sql.NullString{String: req.UserID, Valid: true}
			}

			insertQuery := `
				INSERT INTO device_setting_nodepath 
				(id, device_id, api_key_option, webhook_id, provider, phone_number, api_key, id_device, id_erp, id_admin, instance, created_at, updated_at, user_id)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			`

			_, err = tx.Exec(insertQuery,
				id, deviceID, apiKeyOption, webhookID, provider, phoneNumber, apiKey,
				idDevice, idERP, idAdmin, instance, now, now, userID,
			)

			if err != nil {
				return fmt.Errorf("failed to create device setting: %w", err)
			}

			resultID = id
			logrus.WithFields(logrus.Fields{
				"id":        id,
				"device_id": req.DeviceID,
				"id_device": req.IDDevice,
				"id_erp":    req.IDERP,
				"id_admin":  req.IDAdmin,
			}).Info("Device setting created")

		} else {
			return fmt.Errorf("failed to check existing device setting: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Return the created/updated device setting
	return s.GetByID(resultID)
}

// Create creates a new device setting
func (s *DeviceSettingsService) Create(req *models.CreateDeviceSettingsRequest) (*models.DeviceSettings, error) {
	// Check if database connection is available
	if s.db == nil {
		return nil, fmt.Errorf("database connection is not available")
	}

	id := uuid.New().String()
	now := time.Now()

	// Set defaults if not provided
	apiKeyOption := req.APIKeyOption
	if apiKeyOption == "" {
		apiKeyOption = "openai/gpt-4.1"
	}

	provider := req.Provider
	if provider == "" {
		provider = "wablas"
	}

	// Convert strings to sql.NullString for nullable fields
	var deviceID, webhookID, phoneNumber, apiKey, idDevice, idERP, idAdmin, instance, userID sql.NullString

	if req.DeviceID != "" {
		deviceID = sql.NullString{String: req.DeviceID, Valid: true}
	}
	if req.WebhookID != "" {
		webhookID = sql.NullString{String: req.WebhookID, Valid: true}
	}
	if req.PhoneNumber != "" {
		phoneNumber = sql.NullString{String: req.PhoneNumber, Valid: true}
	}
	if req.APIKey != "" {
		apiKey = sql.NullString{String: req.APIKey, Valid: true}
	}
	if req.IDDevice != "" {
		idDevice = sql.NullString{String: req.IDDevice, Valid: true}
	}
	if req.IDERP != "" {
		idERP = sql.NullString{String: req.IDERP, Valid: true}
	}
	if req.IDAdmin != "" {
		idAdmin = sql.NullString{String: req.IDAdmin, Valid: true}
	}
	if req.Instance != "" {
		instance = sql.NullString{String: req.Instance, Valid: true}
	}
	if req.UserID != "" {
		userID = sql.NullString{String: req.UserID, Valid: true}
	}

	query := `
		INSERT INTO device_setting_nodepath 
		(id, device_id, api_key_option, webhook_id, provider, phone_number, api_key, id_device, id_erp, id_admin, instance, created_at, updated_at, user_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		id,
		deviceID,
		apiKeyOption,
		webhookID,
		provider,
		phoneNumber,
		apiKey,
		idDevice,
		idERP,
		idAdmin,
		instance,
		now,
		now,
		userID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create device setting: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"id":        id,
		"device_id": req.DeviceID,
		"id_device": req.IDDevice,
		"id_erp":    req.IDERP,
		"id_admin":  req.IDAdmin,
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
	// Handle device_id specially - allow empty string to set NULL
	if req.DeviceID == "" {
		existing.DeviceID = sql.NullString{String: "", Valid: false} // Set to NULL for empty string
	} else {
		existing.DeviceID = sql.NullString{String: req.DeviceID, Valid: true}
	}
	if req.APIKeyOption != "" {
		existing.APIKeyOption = req.APIKeyOption
	}
	if req.WebhookID != "" {
		existing.WebhookID = sql.NullString{String: req.WebhookID, Valid: true}
	}
	if req.Provider != "" {
		existing.Provider = req.Provider
	}
	if req.PhoneNumber != "" {
		existing.PhoneNumber = sql.NullString{String: req.PhoneNumber, Valid: true}
	}
	if req.APIKey != "" {
		existing.APIKey = sql.NullString{String: req.APIKey, Valid: true}
	}
	if req.IDDevice != "" {
		existing.IDDevice = sql.NullString{String: req.IDDevice, Valid: true}
	}
	if req.IDERP != "" {
		existing.IDERP = sql.NullString{String: req.IDERP, Valid: true}
	}
	if req.IDAdmin != "" {
		existing.IDAdmin = sql.NullString{String: req.IDAdmin, Valid: true}
	}
	if req.Instance != "" {
		existing.Instance = sql.NullString{String: req.Instance, Valid: true}
	}
	if req.UserID != "" {
		existing.UserID = sql.NullString{String: req.UserID, Valid: true}
	}

	existing.UpdatedAt = time.Now()

	query := `
		UPDATE device_setting_nodepath 
		SET device_id = ?, api_key_option = ?, webhook_id = ?, provider = ?, phone_number = ?, api_key = ?, 
		    id_device = ?, id_erp = ?, id_admin = ?, instance = ?, updated_at = ?, user_id = ?
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
		existing.UserID,
		id,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update device setting: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"id":        id,
		"device_id": existing.DeviceID,
		"id_device": existing.IDDevice,
		"id_erp":    existing.IDERP,
		"id_admin":  existing.IDAdmin,
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
