package repository

import (
	"database/sql"
	"fmt"
	"nodepath-chat/internal/models"
)

type StageSetValueRepository struct {
	db *sql.DB
}

func NewStageSetValueRepository(db *sql.DB) *StageSetValueRepository {
	return &StageSetValueRepository{
		db: db,
	}
}

// GetAll retrieves all stage set values
func (r *StageSetValueRepository) GetAll() ([]*models.StageSetValue, error) {
	if r.db == nil {
		return []*models.StageSetValue{}, nil
	}

	query := `
		SELECT stageSetValue_id, id_device, stage, type_inputData, 
		       columnsData, inputHardCode, created_at, updated_at
		FROM stageSetValue_nodepath
		ORDER BY stageSetValue_id DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query stage set values: %w", err)
	}
	defer rows.Close()

	var values []*models.StageSetValue
	for rows.Next() {
		value := &models.StageSetValue{}
		err := rows.Scan(
			&value.StageSetValueID,
			&value.IDDevice,
			&value.Stage,
			&value.TypeInputData,
			&value.ColumnsData,
			&value.InputHardCode,
			&value.CreatedAt,
			&value.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stage set value: %w", err)
		}
		values = append(values, value)
	}

	return values, nil
}

// GetByDeviceID retrieves all stage set values for a specific device
func (r *StageSetValueRepository) GetByDeviceID(deviceID string) ([]*models.StageSetValue, error) {
	if r.db == nil {
		return []*models.StageSetValue{}, nil
	}

	query := `
		SELECT stageSetValue_id, id_device, stage, type_inputData, 
		       columnsData, inputHardCode, created_at, updated_at
		FROM stageSetValue_nodepath
		WHERE id_device = ?
		ORDER BY stage ASC
	`

	rows, err := r.db.Query(query, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query stage set values by device: %w", err)
	}
	defer rows.Close()

	var values []*models.StageSetValue
	for rows.Next() {
		value := &models.StageSetValue{}
		err := rows.Scan(
			&value.StageSetValueID,
			&value.IDDevice,
			&value.Stage,
			&value.TypeInputData,
			&value.ColumnsData,
			&value.InputHardCode,
			&value.CreatedAt,
			&value.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stage set value: %w", err)
		}
		values = append(values, value)
	}

	return values, nil
}

// GetByStage retrieves a stage set value by device and stage number
func (r *StageSetValueRepository) GetByStage(deviceID string, stage int) (*models.StageSetValue, error) {
	if r.db == nil {
		return nil, sql.ErrNoRows
	}

	query := `
		SELECT stageSetValue_id, id_device, stage, type_inputData, 
		       columnsData, inputHardCode, created_at, updated_at
		FROM stageSetValue_nodepath
		WHERE id_device = ? AND stage = ?
	`

	value := &models.StageSetValue{}
	err := r.db.QueryRow(query, deviceID, stage).Scan(
		&value.StageSetValueID,
		&value.IDDevice,
		&value.Stage,
		&value.TypeInputData,
		&value.ColumnsData,
		&value.InputHardCode,
		&value.CreatedAt,
		&value.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get stage set value: %w", err)
	}

	return value, nil
}

// Create inserts a new stage set value
func (r *StageSetValueRepository) Create(value *models.StageSetValue) error {
	if r.db == nil {
		return fmt.Errorf("database not available")
	}

	query := `
		INSERT INTO stageSetValue_nodepath 
		(id_device, stage, type_inputData, columnsData, inputHardCode, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())
	`

	result, err := r.db.Exec(query,
		value.IDDevice,
		value.Stage,
		value.TypeInputData,
		value.ColumnsData,
		value.InputHardCode,
	)

	if err != nil {
		return fmt.Errorf("failed to create stage set value: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	value.StageSetValueID = int(id)
	return nil
}

// Update updates an existing stage set value
func (r *StageSetValueRepository) Update(value *models.StageSetValue) error {
	if r.db == nil {
		return fmt.Errorf("database not available")
	}

	query := `
		UPDATE stageSetValue_nodepath 
		SET stage = ?, type_inputData = ?, columnsData = ?, 
		    inputHardCode = ?, updated_at = NOW()
		WHERE stageSetValue_id = ?
	`

	_, err := r.db.Exec(query,
		value.Stage,
		value.TypeInputData,
		value.ColumnsData,
		value.InputHardCode,
		value.StageSetValueID,
	)

	if err != nil {
		return fmt.Errorf("failed to update stage set value: %w", err)
	}

	return nil
}

// Delete removes a stage set value by ID
func (r *StageSetValueRepository) Delete(id int) error {
	if r.db == nil {
		return fmt.Errorf("database not available")
	}

	query := `DELETE FROM stageSetValue_nodepath WHERE stageSetValue_id = ?`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete stage set value: %w", err)
	}

	return nil
}
