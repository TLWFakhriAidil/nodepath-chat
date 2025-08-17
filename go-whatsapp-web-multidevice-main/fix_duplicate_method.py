import re

# Read the file
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\repository\broadcast_repository.go', 'r') as f:
    content = f.read()

# Find and remove the duplicate GetPendingMessagesAndLock methods
# First, let's count how many times it appears
count = content.count('func (r *BroadcastRepository) GetPendingMessagesAndLock')
print(f"Found {count} instances of GetPendingMessagesAndLock")

# Remove all instances of the method
pattern = r'// GetPendingMessagesAndLock.*?func \(r \*BroadcastRepository\) GetPendingMessagesAndLock.*?\n}\n'
cleaned = re.sub(pattern, '', content, flags=re.MULTILINE | re.DOTALL)

# Now add back the correct one
new_method = '''// GetPendingMessagesAndLock - MySQL 5.7 compatible version using UPDATE-then-SELECT pattern
func (r *BroadcastRepository) GetPendingMessagesAndLock(deviceID string, limit int) ([]domainBroadcast.BroadcastMessage, error) {
	// Generate unique worker ID for this operation
	workerID := fmt.Sprintf("%s_%d_%s", deviceID, time.Now().UnixNano(), uuid.New().String()[:8])
	
	// MYSQL 5.7 FIX: Use UPDATE-then-SELECT pattern instead of FOR UPDATE SKIP LOCKED
	// Step 1: Atomically claim messages by updating their status and worker ID
	result, err := r.db.Exec(`
		UPDATE broadcast_messages 
		SET status = 'processing',
			processing_worker_id = ?,
			processing_started_at = NOW(),
			updated_at = NOW()
		WHERE device_id = ? 
		AND status = 'pending'
		AND processing_worker_id IS NULL
		AND scheduled_at IS NOT NULL
		AND scheduled_at <= DATE_ADD(NOW(), INTERVAL 8 HOUR)
		AND scheduled_at >= DATE_ADD(DATE_SUB(NOW(), INTERVAL 10 MINUTE), INTERVAL 8 HOUR)
		ORDER BY scheduled_at ASC, group_id, group_order
		LIMIT ?
	`, workerID, deviceID, limit)
	
	if err != nil {
		return nil, fmt.Errorf("failed to claim messages: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// No messages to process
		return []domainBroadcast.BroadcastMessage{}, nil
	}
	
	logrus.Infof("Worker %s claimed %d messages for device %s", workerID, rowsAffected, deviceID)
	
	// Step 2: Fetch the messages we just claimed
	query := `
		SELECT bm.id, bm.user_id, bm.device_id, bm.campaign_id, bm.sequence_id, 
			bm.recipient_phone, bm.recipient_name, bm.message_type, bm.content AS message, bm.media_url, 
			bm.scheduled_at, bm.group_id, bm.group_order, bm.sequence_stepid,
			COALESCE(
				c.min_delay_seconds, 
				ss.min_delay_seconds, 
				s.min_delay_seconds, 
				10
			) AS min_delay,
			COALESCE(
				c.max_delay_seconds, 
				ss.max_delay_seconds, 
				s.max_delay_seconds, 
				30
			) AS max_delay
		FROM broadcast_messages bm
		LEFT JOIN campaigns c ON bm.campaign_id = c.id
		LEFT JOIN sequences s ON bm.sequence_id = s.id
		LEFT JOIN sequence_steps ss ON bm.sequence_stepid = ss.id
		WHERE bm.processing_worker_id = ?
		ORDER BY bm.scheduled_at ASC, bm.group_id, bm.group_order
	`
	
	rows, err := r.db.Query(query, workerID)
	if err != nil {
		// If we can't fetch, reset the messages we claimed
		r.db.Exec(`
			UPDATE broadcast_messages 
			SET status = 'pending',
				processing_worker_id = NULL,
				processing_started_at = NULL
			WHERE processing_worker_id = ?
		`, workerID)
		return nil, err
	}
	defer rows.Close()
	
	var messages []domainBroadcast.BroadcastMessage
	
	for rows.Next() {
		var msg domainBroadcast.BroadcastMessage
		var userID sql.NullString
		var campaignID sql.NullInt64
		var sequenceID, groupID, sequenceStepID sql.NullString
		var groupOrder sql.NullInt64
		var scheduledAt sql.NullTime
		
		err := rows.Scan(&msg.ID, &userID, &msg.DeviceID, &campaignID, &sequenceID,
			&msg.RecipientPhone, &msg.RecipientName, &msg.Type, &msg.Content, &msg.MediaURL, &scheduledAt,
			&groupID, &groupOrder, &sequenceStepID, &msg.MinDelay, &msg.MaxDelay)
		if err != nil {
			logrus.Warnf("Error scanning message: %v", err)
			continue
		}
		
		if userID.Valid {
			msg.UserID = userID.String
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
		if groupID.Valid {
			msg.GroupID = &groupID.String
		}
		if groupOrder.Valid {
			groupOrderInt := int(groupOrder.Int64)
			msg.GroupOrder = &groupOrderInt
		}
		if scheduledAt.Valid {
			msg.ScheduledAt = scheduledAt.Time
		}
		
		// Set ImageURL for backward compatibility
		msg.ImageURL = msg.MediaURL
		msg.Message = msg.Content
		
		messages = append(messages, msg)
	}
	
	logrus.Infof("Worker %s fetched %d messages for device %s", workerID, len(messages), deviceID)
	
	return messages, nil
}
'''

# Add the method at the end of the file (before the last closing brace if there is one)
if cleaned.rstrip().endswith('}'):
    cleaned = cleaned.rstrip()[:-1] + '\n' + new_method + '\n}'
else:
    cleaned = cleaned.rstrip() + '\n\n' + new_method

# Write back
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\repository\broadcast_repository.go', 'w') as f:
    f.write(cleaned)

print("Fixed the duplicate method issue!")