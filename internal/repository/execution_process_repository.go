package repository

import (
	"database/sql"
	"time"

	"nodepath-chat/internal/models"

	"github.com/sirupsen/logrus"
)

// ExecutionProcessRepository handles database operations for execution process tracking
type ExecutionProcessRepository interface {
	CreateExecution(idDevice, idProspect string) (int, error)
	GetOldestExecution(idDevice, idProspect string) (*models.ExecutionProcess, error)
	DeleteExecutions(idDevice, idProspect string) error
}

type executionProcessRepository struct {
	db *sql.DB
}

// NewExecutionProcessRepository creates a new execution process repository
func NewExecutionProcessRepository(db *sql.DB) ExecutionProcessRepository {
	return &executionProcessRepository{db: db}
}

// CreateExecution creates a new execution record and returns its ID
func (r *executionProcessRepository) CreateExecution(idDevice, idProspect string) (int, error) {
	query := `
		INSERT INTO execution_process_nodepath (id_device, id_prospect, times)
		VALUES (?, ?, ?)
	`

	result, err := r.db.Exec(query, idDevice, idProspect, time.Now())
	if err != nil {
		logrus.WithError(err).Error("Failed to create execution record")
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		logrus.WithError(err).Error("Failed to get last insert ID")
		return 0, err
	}

	logrus.WithFields(logrus.Fields{
		"id_device":    idDevice,
		"id_prospect":  idProspect,
		"id_execution": id,
	}).Info("✅ Created execution record")

	return int(id), nil
}

// GetOldestExecution gets the oldest execution record for a device and prospect
func (r *executionProcessRepository) GetOldestExecution(idDevice, idProspect string) (*models.ExecutionProcess, error) {
	query := `
		SELECT id_chatInput, id_device, id_prospect, times
		FROM execution_process_nodepath
		WHERE id_device = ? AND id_prospect = ?
		ORDER BY id_chatInput ASC
		LIMIT 1
	`

	var exec models.ExecutionProcess
	err := r.db.QueryRow(query, idDevice, idProspect).Scan(
		&exec.IDChatInput,
		&exec.IDDevice,
		&exec.IDProspect,
		&exec.Times,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		logrus.WithError(err).Error("Failed to get oldest execution record")
		return nil, err
	}

	return &exec, nil
}

// DeleteExecutions deletes all execution records for a device and prospect
func (r *executionProcessRepository) DeleteExecutions(idDevice, idProspect string) error {
	query := `
		DELETE FROM execution_process_nodepath
		WHERE id_device = ? AND id_prospect = ?
	`

	result, err := r.db.Exec(query, idDevice, idProspect)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete execution records")
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	logrus.WithFields(logrus.Fields{
		"id_device":     idDevice,
		"id_prospect":   idProspect,
		"rows_affected": rowsAffected,
	}).Info("🧹 Cleaned up execution records")

	return nil
}
