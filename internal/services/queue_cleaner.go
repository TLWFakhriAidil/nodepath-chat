package services

import (
	"context"
	"time"

	"nodepath-chat/internal/repository"

	"github.com/sirupsen/logrus"
)

// QueueCleaner handles cleanup of stuck and timed-out messages
type QueueCleaner struct {
	ctx    context.Context
	cancel context.CancelFunc
	repo   *repository.BroadcastRepository
}

// NewQueueCleaner creates a new queue cleaner service
func NewQueueCleaner() *QueueCleaner {
	ctx, cancel := context.WithCancel(context.Background())
	return &QueueCleaner{
		ctx:    ctx,
		cancel: cancel,
		repo:   repository.GetBroadcastRepository(),
	}
}

// Start begins the queue cleaning process
func (qc *QueueCleaner) Start() {
	logrus.Info("Starting queue cleaner service")
	go qc.cleanupLoop()
}

// Stop gracefully shuts down the queue cleaner
func (qc *QueueCleaner) Stop() {
	logrus.Info("Stopping queue cleaner service")
	qc.cancel()
}

// cleanupLoop runs the cleanup process periodically
func (qc *QueueCleaner) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute) // Run every 5 minutes
	defer ticker.Stop()
	
	for {
		select {
		case <-qc.ctx.Done():
			logrus.Info("Queue cleaner shutdown")
			return
		case <-ticker.C:
			qc.performCleanup()
		}
	}
}

// performCleanup executes the actual cleanup operations
func (qc *QueueCleaner) performCleanup() {
	logrus.Debug("Starting queue cleanup cycle")
	
	// Reset stuck 'queued' messages to 'pending'
	// Messages stuck in 'queued' status for more than 10 minutes
	stuckThreshold := time.Now().Add(-10 * time.Minute)
	stuckCount, err := qc.resetStuckQueuedMessages(stuckThreshold)
	if err != nil {
		logrus.Errorf("Failed to reset stuck queued messages: %v", err)
	} else if stuckCount > 0 {
		logrus.Infof("Reset %d stuck queued messages to pending", stuckCount)
	}
	
	// Reset stuck 'processing' messages to 'pending'
	// Messages stuck in 'processing' status for more than 15 minutes
	processingThreshold := time.Now().Add(-15 * time.Minute)
	processingCount, err := qc.resetStuckProcessingMessages(processingThreshold)
	if err != nil {
		logrus.Errorf("Failed to reset stuck processing messages: %v", err)
	} else if processingCount > 0 {
		logrus.Infof("Reset %d stuck processing messages to pending", processingCount)
	}
	
	// Mark old messages as failed
	// Messages older than 12 hours that are still pending/queued/processing
	failThreshold := time.Now().Add(-12 * time.Hour)
	failedCount, err := qc.markOldMessagesAsFailed(failThreshold)
	if err != nil {
		logrus.Errorf("Failed to mark old messages as failed: %v", err)
	} else if failedCount > 0 {
		logrus.Infof("Marked %d old messages as failed", failedCount)
	}
	
	// Clean up very old failed messages (optional)
	// Remove failed messages older than 7 days to prevent database bloat
	cleanupThreshold := time.Now().Add(-7 * 24 * time.Hour)
	cleanedCount, err := qc.cleanupOldFailedMessages(cleanupThreshold)
	if err != nil {
		logrus.Errorf("Failed to cleanup old failed messages: %v", err)
	} else if cleanedCount > 0 {
		logrus.Infof("Cleaned up %d old failed messages", cleanedCount)
	}
	
	logrus.Debug("Queue cleanup cycle completed")
}

// resetStuckQueuedMessages resets messages stuck in 'queued' status
func (qc *QueueCleaner) resetStuckQueuedMessages(threshold time.Time) (int, error) {
	query := `
		UPDATE broadcast_messages 
		SET status = 'pending', 
		    updated_at = NOW(),
		    error_message = 'Reset from stuck queued status'
		WHERE status = 'queued' 
		  AND updated_at < ?
	`
	
	result, err := qc.repo.DB().Exec(query, threshold)
	if err != nil {
		return 0, err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	
	return int(rowsAffected), nil
}

// resetStuckProcessingMessages resets messages stuck in 'processing' status
func (qc *QueueCleaner) resetStuckProcessingMessages(threshold time.Time) (int, error) {
	query := `
		UPDATE broadcast_messages 
		SET status = 'pending', 
		    updated_at = NOW(),
		    error_message = 'Reset from stuck processing status'
		WHERE status = 'processing' 
		  AND updated_at < ?
	`
	
	result, err := qc.repo.DB().Exec(query, threshold)
	if err != nil {
		return 0, err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	
	return int(rowsAffected), nil
}

// markOldMessagesAsFailed marks very old messages as failed
func (qc *QueueCleaner) markOldMessagesAsFailed(threshold time.Time) (int, error) {
	query := `
		UPDATE broadcast_messages 
		SET status = 'failed', 
		    updated_at = NOW(),
		    error_message = 'Message timeout - exceeded maximum processing time'
		WHERE status IN ('pending', 'queued', 'processing') 
		  AND created_at < ?
	`
	
	result, err := qc.repo.DB().Exec(query, threshold)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rowsAffected), nil
}

// cleanupOldFailedMessages removes very old failed messages
func (qc *QueueCleaner) cleanupOldFailedMessages(threshold time.Time) (int, error) {
	query := `
		DELETE FROM broadcast_messages 
		WHERE status = 'failed' 
		  AND updated_at < ?
	`

	result, err := qc.repo.DB().Exec(query, threshold)
	if err != nil {
		return 0, err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	
	return int(rowsAffected), nil
}

// GetQueueStats returns current queue statistics
func (qc *QueueCleaner) GetQueueStats() (map[string]int, error) {
	stats := make(map[string]int)
	
	// Count messages by status
	statusQuery := `
		SELECT status, COUNT(*) as count 
		FROM broadcast_messages 
		GROUP BY status
	`
	
	rows, err := qc.repo.DB().Query(statusQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats[status] = count
	}
	
	// Count stuck messages
	stuckQueuedQuery := `
		SELECT COUNT(*) 
		FROM broadcast_messages 
		WHERE status = 'queued' 
		  AND updated_at < ?
	`
	
	var stuckQueued int
	err = qc.repo.DB().QueryRow(stuckQueuedQuery, time.Now().Add(-10*time.Minute)).Scan(&stuckQueued)
	if err != nil {
		return nil, err
	}
	stats["stuck_queued"] = stuckQueued
	
	stuckProcessingQuery := `
		SELECT COUNT(*) 
		FROM broadcast_messages 
		WHERE status = 'processing' 
		  AND updated_at < ?
	`
	
	var stuckProcessing int
	err = qc.repo.DB().QueryRow(stuckProcessingQuery, time.Now().Add(-15*time.Minute)).Scan(&stuckProcessing)
	if err != nil {
		return nil, err
	}
	stats["stuck_processing"] = stuckProcessing
	
	return stats, nil
}

// ForceCleanup manually triggers a cleanup cycle
func (qc *QueueCleaner) ForceCleanup() {
	logrus.Info("Manual queue cleanup triggered")
	qc.performCleanup()
}