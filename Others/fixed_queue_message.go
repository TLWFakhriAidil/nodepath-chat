// QueueMessage adds a message to the queue with proper duplicate prevention
func (r *BroadcastRepository) QueueMessage(msg domainBroadcast.BroadcastMessage) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	
	// Use a transaction for atomic duplicate check and insert
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// ISSUE 3 FIX: Check for duplicates before inserting
	// For SEQUENCES: Check based on sequence_stepid, recipient_phone, and device_id
	if msg.SequenceStepID != nil && *msg.SequenceStepID != "" {
		duplicateCheck := `
			SELECT COUNT(*) 
			FROM broadcast_messages 
			WHERE sequence_stepid = ? 
			AND recipient_phone = ? 
			AND device_id = ?
			AND status IN ('pending', 'sent', 'queued', 'processing')
			FOR UPDATE
		`
		
		var count int
		err := tx.QueryRow(duplicateCheck, *msg.SequenceStepID, msg.RecipientPhone, msg.DeviceID).Scan(&count)
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
			FOR UPDATE
		`
		
		var count int
		err := tx.QueryRow(duplicateCheck, *msg.CampaignID, msg.RecipientPhone, msg.DeviceID).Scan(&count)
		if err != nil {
			logrus.Warnf("Error checking campaign duplicates: %v", err)
		} else if count > 0 {
			logrus.Infof("Skipping duplicate campaign message for %s - campaign %d already exists", 
				msg.RecipientPhone, *msg.CampaignID)
			return nil // Skip duplicate
		}
	}
	
	query := `
		INSERT INTO broadcast_messages(id, user_id, device_id, campaign_id, sequence_id, sequence_stepid, recipient_phone, recipient_name,
		 message_type, content, media_url, status, scheduled_at, created_at, group_id, group_order)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`	
	// Get user_id - use from message if provided, otherwise get from device
	var userID string
	if msg.UserID != "" {
		userID = msg.UserID
	} else {
		err := tx.QueryRow("SELECT user_id from user_devices WHERE id = ?", msg.DeviceID).Scan(&userID)
		if err != nil {
			return err
		}
	}
	
	// Handle nullable fields
	var campaignID interface{}
	if msg.CampaignID != nil {
		campaignID = *msg.CampaignID
	} else {
		campaignID = nil
	}
	
	var sequenceID interface{}
	if msg.SequenceID != nil && *msg.SequenceID != "" {
		sequenceID = *msg.SequenceID
	} else {
		sequenceID = nil
	}
	
	var sequenceStepID interface{}
	if msg.SequenceStepID != nil && *msg.SequenceStepID != "" {
		sequenceStepID = *msg.SequenceStepID
	} else {
		sequenceStepID = nil
	}
	
	var groupID interface{}
	if msg.GroupID != nil && *msg.GroupID != "" {
		groupID = *msg.GroupID
	} else {
		groupID = nil
	}	
	var groupOrder interface{}
	if msg.GroupOrder != nil {
		groupOrder = *msg.GroupOrder
	} else {
		groupOrder = nil
	}
	
	_, err = tx.Exec(query, msg.ID, userID, msg.DeviceID, campaignID,
		sequenceID, sequenceStepID, msg.RecipientPhone, msg.RecipientName, msg.Type, msg.Content, 
		msg.MediaURL, "pending", msg.ScheduledAt, time.Now(), groupID, groupOrder)
	
	if err != nil {
		return err
	}
	
	// Commit the transaction
	return tx.Commit()
}
