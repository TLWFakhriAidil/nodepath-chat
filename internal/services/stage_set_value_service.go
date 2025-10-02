package services

import (
	"database/sql"
	"fmt"
	"nodepath-chat/internal/models"
	"nodepath-chat/internal/repository"
)

type StageSetValueService struct {
	repo *repository.StageSetValueRepository
}

func NewStageSetValueService(db *sql.DB) *StageSetValueService {
	return &StageSetValueService{
		repo: repository.NewStageSetValueRepository(db),
	}
}

// GetAll retrieves all stage set values
func (s *StageSetValueService) GetAll() ([]*models.StageSetValue, error) {
	return s.repo.GetAll()
}

// GetByDeviceID retrieves all stage set values for a specific device
func (s *StageSetValueService) GetByDeviceID(deviceID string) ([]*models.StageSetValue, error) {
	return s.repo.GetByDeviceID(deviceID)
}

// GetByStage retrieves a stage set value by device and stage number
func (s *StageSetValueService) GetByStage(deviceID string, stage int) (*models.StageSetValue, error) {
	return s.repo.GetByStage(deviceID, stage)
}

// Create creates a new stage set value
func (s *StageSetValueService) Create(value *models.StageSetValue) error {
	// Validate required fields
	if value.IDDevice == "" {
		return fmt.Errorf("device ID is required")
	}

	if value.Stage <= 0 {
		return fmt.Errorf("stage must be a positive number")
	}

	if value.TypeInputData != "User Input" && value.TypeInputData != "Set" {
		return fmt.Errorf("type must be 'User Input' or 'Set'")
	}

	if value.ColumnsData == "" {
		return fmt.Errorf("column is required")
	}

	// Validate column name
	validColumns := map[string]bool{
		"nama":        true,
		"alamat":      true,
		"pakej":       true,
		"no_fon":      true,
		"tarikh_gaji": true,
	}

	if !validColumns[value.ColumnsData] {
		return fmt.Errorf("invalid column: %s", value.ColumnsData)
	}

	// If type is Set, inputHardCode is required
	if value.TypeInputData == "Set" && (!value.InputHardCode.Valid || value.InputHardCode.String == "") {
		return fmt.Errorf("hardcoded value is required when type is 'Set'")
	}

	// If type is User Input, clear inputHardCode
	if value.TypeInputData == "User Input" {
		value.InputHardCode = sql.NullString{Valid: false}
	}

	return s.repo.Create(value)
}

// Update updates an existing stage set value
func (s *StageSetValueService) Update(value *models.StageSetValue) error {
	// Validate required fields
	if value.Stage <= 0 {
		return fmt.Errorf("stage must be a positive number")
	}

	if value.TypeInputData != "User Input" && value.TypeInputData != "Set" {
		return fmt.Errorf("type must be 'User Input' or 'Set'")
	}

	if value.ColumnsData == "" {
		return fmt.Errorf("column is required")
	}

	// Validate column name
	validColumns := map[string]bool{
		"nama":        true,
		"alamat":      true,
		"pakej":       true,
		"no_fon":      true,
		"tarikh_gaji": true,
	}

	if !validColumns[value.ColumnsData] {
		return fmt.Errorf("invalid column: %s", value.ColumnsData)
	}

	// If type is Set, inputHardCode is required
	if value.TypeInputData == "Set" && (!value.InputHardCode.Valid || value.InputHardCode.String == "") {
		return fmt.Errorf("hardcoded value is required when type is 'Set'")
	}

	// If type is User Input, clear inputHardCode
	if value.TypeInputData == "User Input" {
		value.InputHardCode = sql.NullString{Valid: false}
	}

	return s.repo.Update(value)
}

// Delete deletes a stage set value by ID
func (s *StageSetValueService) Delete(id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid ID")
	}
	return s.repo.Delete(id)
}

// ProcessStageValue applies the stage set value logic to save data
func (s *StageSetValueService) ProcessStageValue(deviceID string, stage int, userInput string) (map[string]interface{}, error) {
	updates := make(map[string]interface{})

	// Get the stage set value configuration
	config, err := s.GetByStage(deviceID, stage)
	if err != nil {
		// No configuration for this stage, return empty updates
		if err == sql.ErrNoRows {
			return updates, nil
		}
		return nil, fmt.Errorf("failed to get stage configuration: %w", err)
	}

	// Determine the value to save
	var valueToSave string
	if config.TypeInputData == "Set" && config.InputHardCode.Valid {
		// Use the hardcoded value
		valueToSave = config.InputHardCode.String
	} else {
		// Use the user input
		valueToSave = userInput
	}

	// Add the update for the specified column
	updates[config.ColumnsData] = valueToSave

	return updates, nil
}
