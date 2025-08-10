package repository

import (
	"database/sql"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/database"
)

type workerRepository struct {
	db *sql.DB
}

var workerRepo *workerRepository

// GetWorkerRepository returns worker repository instance
func GetWorkerRepository() *workerRepository {
	if workerRepo == nil {
		workerRepo = &workerRepository{
			db: database.GetDB(),
		}
	}
	return workerRepo
}

// UpdateWorkerStatus updates or inserts worker status
func (r *workerRepository) UpdateWorkerStatus(deviceID, status string, queueSize int, processed, failed int64) error {
	query := `
		INSERT INTO worker_status(device_id, worker_type, ` + "`status`" + `, current_queue_size, messages_processed, messages_failed, last_activity, updated_at)
		VALUES (?, 'broadcast', ?, ?, ?, ?, ?, ?)
		ON CONFLICT (device_id, worker_type) 
		DO UPDATE SET 
			status = ?,
			current_queue_size = ?,
			messages_processed = ?,
			messages_failed = ?,
			last_activity = ?,
			updated_at = ?
	`
	
	now := time.Now()
	_, err := r.db.Exec(query, deviceID, status, queueSize, processed, failed, now, now)
	return err
}

// GetWorkerStatuses gets all worker statuses
func (r *workerRepository) GetWorkerStatuses() ([]map[string]interface{}, error) {
	query := `
		SELECT device_id, worker_type, ` + "`status`" + `, current_queue_size, 
		       messages_processed, messages_failed, last_activity
		FROM worker_status
		ORDER BY last_activity DESC
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var statuses []map[string]interface{}
	for rows.Next() {
		var deviceID, workerType, status string
		var queueSize int
		var processed, failed int64
		var lastActivity time.Time
		
		err := rows.Scan(&deviceID, &workerType, &status, &queueSize, 
			&processed, &failed, &lastActivity)
		if err != nil {
			continue
		}
		
		statuses = append(statuses, map[string]interface{}{
			"device_id":          deviceID,
			"worker_type":        workerType,
			"status":             status,
			"queue_size":         queueSize,
			"messages_processed": processed,
			"messages_failed":    failed,
			"last_activity":      lastActivity,
		})
	}
	
	return statuses, nil
}

// CleanupOldWorkers removes old worker records
func (r *workerRepository) CleanupOldWorkers(olderThan time.Duration) error {
	query := `DELETE FROM worker_status WHERE last_activity < ?`
	cutoff := time.Now().Add(-olderThan)
	_, err := r.db.Exec(query, cutoff)
	return err
}
