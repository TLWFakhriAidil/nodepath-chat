package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"nodepath-chat/internal/database"
	"nodepath-chat/internal/models"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type BroadcastRepository struct {
	db *sql.DB
}

var broadcastRepo *BroadcastRepository

// GetBroadcastRepository returns broadcast repository instance
func GetBroadcastRepository() *BroadcastRepository {
	if broadcastRepo == nil {
		broadcastRepo = &BroadcastRepository{
			db: database.GetDB(),
		}
	}
	return broadcastRepo
}

// QueueMessage adds a message to the broadcast queue with duplicate prevention
func (r *BroadcastRepository) QueueMessage(msg models.BroadcastMessage) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	
	// Check for duplicates before inserting
	// For SEQUENCES: Check based on sequence_stepid, recipient_phone, and device_id
	if msg.SequenceStepID != nil && *msg.SequenceStepID != "" {
		duplicateCheck := `
			SELECT COUNT(*) 
			FROM broadcast_messages 
			WHERE sequence_stepid = ? 
			AND recipient_phone = ? 
			AND device_id = ?
			AND status IN ('pending', 'sent', 'queued', 'processing')
		`
		
		var count int
		err := r.db.QueryRow(duplicateCheck, *msg.SequenceStepID, msg.RecipientPhone, msg.DeviceID).Scan(&count)
		if err != nil {
			logrus.Warnf("Error checking sequence duplicates: %v", err)
		} else if count > 0 {
			logrus.Infof("Skipping duplicate sequence message for %s - sequence_step %s already exists", 
				msg.RecipientPhone, *msg.SequenceStepID)
			return nil // Skip duplicate
		}
	}
	
	// For CAMPAIGNS: Check based on campaign_id, recipient_phone, and device_id
	if msg.CampaignID != nil && *msg.CampaignID > 0 {
		duplicateCheck := `
			SELECT COUNT(*) 
			FROM broadcast_messages 
			WHERE campaign_id = ? 
			AND recipient_phone = ? 
			AND device_id = ?
			AND status IN ('pending', 'sent', 'queued', 'processing')
		`
		
		var count int
		err := r.db.QueryRow(duplicateCheck, *msg.CampaignID, msg.RecipientPhone, msg.DeviceID).Scan(&count)
		if err != nil {
			logrus.Warnf("Error checking campaign duplicates: %v", err)
		} else if count > 0 {
			logrus.Infof("Skipping duplicate campaign message for %s - campaign %d already exists", 
				msg.RecipientPhone, *msg.CampaignID)
			return nil // Skip duplicate
		}
	}
	
	// Insert the message
	query := `
		INSERT INTO broadcast_messages (
			id, user_id, device_id, campaign_id, sequence_id, sequence_stepid,
			recipient_phone, recipient_name, message_type, content, media_url,
			status, scheduled_at, created_at, updated_at, group_id, group_order
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	now := time.Now()
	scheduledAt := msg.ScheduledAt
	if scheduledAt.IsZero() {
		scheduledAt = now
	}
	
	_, err := r.db.Exec(query,
		msg.ID, msg.UserID, msg.DeviceID, msg.CampaignID, msg.SequenceID, msg.SequenceStepID,
		msg.RecipientPhone, msg.RecipientName, msg.Type, msg.Content, msg.MediaURL,
		msg.Status, scheduledAt, now, now, msg.GroupID, msg.GroupOrder)
	
	if err != nil {
		return fmt.Errorf("failed to insert broadcast message: %w", err)
	}
	
	logrus.Debugf("Queued message %s for %s to device %s", msg.ID, msg.RecipientPhone, msg.DeviceID)
	return nil
}

// GetPendingMessages gets pending messages for a device with optimized query
func (r *BroadcastRepository) GetPendingMessages(deviceID string, limit int) ([]models.BroadcastMessage, error) {
	query := `
		SELECT bm.id, bm.user_id, bm.device_id, bm.campaign_id, bm.sequence_id, bm.sequence_stepid,
			bm.recipient_phone, bm.recipient_name, bm.message_type, bm.content, bm.media_url, 
			bm.scheduled_at, bm.group_id, bm.group_order,
			COALESCE(
				c.min_delay_seconds, 
				ss.min_delay_seconds, 
				s.min_delay_seconds, 
				ud.min_delay_seconds,
				5
			) AS min_delay,
			COALESCE(
				c.max_delay_seconds, 
				ss.max_delay_seconds, 
				s.max_delay_seconds, 
				ud.max_delay_seconds,
				15
			) AS max_delay
		FROM broadcast_messages bm
		LEFT JOIN campaigns c ON bm.campaign_id = c.id
		LEFT JOIN sequences s ON bm.sequence_id = s.id
		LEFT JOIN sequence_steps ss ON bm.sequence_stepid = ss.id
		LEFT JOIN user_devices ud ON bm.device_id = ud.id
		WHERE bm.device_id = ? 
		AND bm.status = 'pending'
		AND (bm.scheduled_at IS NULL OR bm.scheduled_at <= NOW())
		ORDER BY bm.scheduled_at ASC, bm.group_id, bm.group_order
		LIMIT ?
	`	
	
	rows, err := r.db.Query(query, deviceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var messages []models.BroadcastMessage
	for rows.Next() {
		var msg models.BroadcastMessage
		var userID sql.NullString
		var campaignID sql.NullInt64
		var sequenceID sql.NullString
		var sequenceStepID sql.NullString
		var recipientName sql.NullString
		var mediaURL sql.NullString
		var scheduledAt sql.NullTime
		var groupID sql.NullString
		var groupOrder sql.NullInt64
		var minDelay, maxDelay int
		
		err := rows.Scan(&msg.ID, &userID, &msg.DeviceID, &campaignID, &sequenceID, &sequenceStepID,
			&msg.RecipientPhone, &recipientName, &msg.Type, &msg.Content, &mediaURL, 
			&scheduledAt, &groupID, &groupOrder, &minDelay, &maxDelay)
		if err != nil {
			continue
		}
		
		// Set nullable fields
		if userID.Valid {
			msg.UserID = userID.String
		}
		if recipientName.Valid {
			msg.RecipientName = recipientName.String
		}
		if mediaURL.Valid {
			msg.MediaURL = mediaURL.String
		}
		if campaignID.Valid {
			campaignIDInt := int(campaignID.Int64)
			msg.CampaignID = &campaignIDInt
		}
		if sequenceID.Valid {
			msg.SequenceID = &sequenceID.String
		}
		if sequenceStepID.Valid {
			msg.SequenceStepID = &sequenceStepID.String
		}
		if scheduledAt.Valid {
			msg.ScheduledAt = scheduledAt.Time
		}
		if groupID.Valid {
			msg.GroupID = &groupID.String
		}
		if groupOrder.Valid {
			groupOrderInt := int(groupOrder.Int64)
			msg.GroupOrder = &groupOrderInt
		}
		
		// Set delay values
		msg.MinDelay = minDelay
		msg.MaxDelay = maxDelay
		
		messages = append(messages, msg)
	}
	
	return messages, rows.Err()
}

// GetAllPendingMessages gets all pending messages across all devices for batch processing
func (r *BroadcastRepository) GetAllPendingMessages(limit int) ([]models.BroadcastMessage, error) {
	query := `
		SELECT id, user_id, device_id, campaign_id, sequence_id, sequence_stepid,
		       recipient_phone, recipient_name, message_type, content, media_url, 
		       status, scheduled_at, created_at, group_id, group_order
		FROM broadcast_messages
		WHERE status = 'pending' 
		AND (scheduled_at IS NULL OR scheduled_at <= NOW())
		ORDER BY scheduled_at ASC, created_at ASC
		LIMIT ?
	`
	
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var messages []models.BroadcastMessage
	for rows.Next() {
		var msg models.BroadcastMessage
		var userID sql.NullString
		var campaignID sql.NullInt64
		var sequenceID sql.NullString
		var sequenceStepID sql.NullString
		var recipientName sql.NullString
		var mediaURL sql.NullString
		var scheduledAt sql.NullTime
		var groupID sql.NullString
		var groupOrder sql.NullInt64
		
		err := rows.Scan(&msg.ID, &userID, &msg.DeviceID, &campaignID, &sequenceID, &sequenceStepID,
			&msg.RecipientPhone, &recipientName, &msg.Type, &msg.Content, &mediaURL, 
			&msg.Status, &scheduledAt, &msg.CreatedAt, &groupID, &groupOrder)
		if err != nil {
			continue
		}
		
		// Set nullable fields
		if userID.Valid {
			msg.UserID = userID.String
		}
		if recipientName.Valid {
			msg.RecipientName = recipientName.String
		}
		if mediaURL.Valid {
			msg.MediaURL = mediaURL.String
		}
		if campaignID.Valid {
			campaignIDInt := int(campaignID.Int64)
			msg.CampaignID = &campaignIDInt
		}
		if sequenceID.Valid {
			msg.SequenceID = &sequenceID.String
		}
		if sequenceStepID.Valid {
			msg.SequenceStepID = &sequenceStepID.String
		}
		if scheduledAt.Valid {
			msg.ScheduledAt = scheduledAt.Time
		}
		if groupID.Valid {
			msg.GroupID = &groupID.String
		}
		if groupOrder.Valid {
			groupOrderInt := int(groupOrder.Int64)
			msg.GroupOrder = &groupOrderInt
		}
		
		messages = append(messages, msg)
	}
	
	return messages, rows.Err()
}

// UpdateMessageStatus updates the status of a broadcast message
func (r *BroadcastRepository) UpdateMessageStatus(messageID, status, errorMessage string) error {
	query := `
		UPDATE broadcast_messages 
		SET status = ?, error_message = ?, updated_at = ?
		WHERE id = ?
	`
	
	_, err := r.db.Exec(query, status, errorMessage, time.Now(), messageID)
	if err != nil {
		return fmt.Errorf("failed to update message status: %w", err)
	}
	
	return nil
}

// GetDevicesWithPendingMessages gets all device IDs that have pending messages
func (r *BroadcastRepository) GetDevicesWithPendingMessages() ([]string, error) {
	query := `
		SELECT DISTINCT device_id 
		FROM broadcast_messages 
		WHERE status = 'pending' 
		AND (scheduled_at IS NULL OR scheduled_at <= NOW())
		ORDER BY device_id
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var devices []string
	for rows.Next() {
		var deviceID string
		if err := rows.Scan(&deviceID); err != nil {
			return nil, err
		}
		devices = append(devices, deviceID)
	}
	
	return devices, rows.Err()
}

// BatchUpdateMessageStatus updates multiple message statuses in a single transaction
func (r *BroadcastRepository) BatchUpdateMessageStatus(messageIDs []string, status string, errorMessage string) error {
	if len(messageIDs) == 0 {
		return nil
	}
	
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	query := `UPDATE broadcast_messages SET status = ?, error_message = ?, updated_at = ? WHERE id = ?`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()
	
	now := time.Now()
	for _, messageID := range messageIDs {
		_, err = stmt.Exec(status, errorMessage, now, messageID)
		if err != nil {
			return fmt.Errorf("failed to update message %s: %w", messageID, err)
		}
	}
	
	return tx.Commit()
}

// QueueBulkMessages queues multiple messages in a single transaction
func (r *BroadcastRepository) QueueBulkMessages(messages []models.BroadcastMessage) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	query := `
		INSERT INTO broadcast_messages (
			id, user_id, device_id, campaign_id, sequence_id, sequence_stepid,
			recipient_phone, recipient_name, message_type, content, media_url,
			status, scheduled_at, created_at, updated_at, group_id, group_order
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()
	
	for _, msg := range messages {
		now := time.Now()
		scheduledAt := msg.ScheduledAt
		if scheduledAt.IsZero() {
			scheduledAt = now
		}
		
		_, err = stmt.Exec(
			msg.ID, msg.UserID, msg.DeviceID, msg.CampaignID, msg.SequenceID, msg.SequenceStepID,
			msg.RecipientPhone, msg.RecipientName, msg.Type, msg.Content, msg.MediaURL,
			msg.Status, scheduledAt, now, now, msg.GroupID, msg.GroupOrder,
		)
		if err != nil {
			return fmt.Errorf("failed to insert message %s: %w", msg.ID, err)
		}
	}
	
	return tx.Commit()
}

// GetMessageStatus returns the status of a specific message
func (r *BroadcastRepository) GetMessageStatus(messageID string) (string, error) {
	var status string
	query := `SELECT status FROM broadcast_messages WHERE id = ?`
	err := r.db.QueryRow(query, messageID).Scan(&status)
	if err != nil {
		return "", fmt.Errorf("failed to get message status: %w", err)
	}
	return status, nil
}

// GetPendingMessageCount returns the number of pending messages for a device
func (r *BroadcastRepository) GetPendingMessageCount(deviceID string) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM broadcast_messages 
		WHERE device_id = ? AND status = 'pending' AND (scheduled_at IS NULL OR scheduled_at <= NOW())
	`
	err := r.db.QueryRow(query, deviceID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get pending message count: %w", err)
	}
	return count, nil
}

// GetTotalPendingMessageCount returns the total number of pending messages across all devices
func (r *BroadcastRepository) GetTotalPendingMessageCount() (int, error) {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM broadcast_messages 
		WHERE status = 'pending' AND (scheduled_at IS NULL OR scheduled_at <= NOW())
	`
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get total pending message count: %w", err)
	}
	return count, nil
}

// CancelCampaignMessages cancels all pending messages for a campaign
func (r *BroadcastRepository) CancelCampaignMessages(campaignID int) (int, error) {
	query := `
		UPDATE broadcast_messages 
		SET status = 'cancelled', updated_at = ?, error_message = 'Cancelled by user'
		WHERE campaign_id = ? AND status IN ('pending', 'queued')
	`
	result, err := r.db.Exec(query, time.Now(), campaignID)
	if err != nil {
		return 0, fmt.Errorf("failed to cancel campaign messages: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	return int(rowsAffected), nil
}

// CancelSequenceMessages cancels all pending messages for a sequence
func (r *BroadcastRepository) CancelSequenceMessages(sequenceID string) (int, error) {
	query := `
		UPDATE broadcast_messages 
		SET status = 'cancelled', updated_at = ?, error_message = 'Cancelled by user'
		WHERE sequence_id = ? AND status IN ('pending', 'queued')
	`
	result, err := r.db.Exec(query, time.Now(), sequenceID)
	if err != nil {
		return 0, fmt.Errorf("failed to cancel sequence messages: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	return int(rowsAffected), nil
}