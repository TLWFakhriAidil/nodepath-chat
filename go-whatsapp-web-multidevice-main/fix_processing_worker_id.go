// Fix for GetPendingMessagesAndLock in broadcast_repository.go
// Replace the update section (around line 520-544) with this:

	// If we got messages, update their status to processing
	if len(messageIDs) > 0 {
		placeholders := make([]string, len(messageIDs))
		args := make([]interface{}, 0, len(messageIDs)+1)
		args = append(args, workerID)
		
		// Add debug logging
		logrus.Infof("Setting processing_worker_id to: %s for %d messages", workerID, len(messageIDs))
		
		for i, id := range messageIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		
		updateQuery := fmt.Sprintf(`
			UPDATE broadcast_messages 
			SET status = 'processing',
				processing_worker_id = ?,
				processing_started_at = NOW(),
				updated_at = NOW()
			WHERE id IN (%s)
		`, strings.Join(placeholders, ","))
		
		// Log the query for debugging
		logrus.Debugf("Update query: %s", updateQuery)
		logrus.Debugf("Args count: %d (workerID + %d message IDs)", len(args), len(messageIDs))
		
		result, err := tx.Exec(updateQuery, args...)
		if err != nil {
			// This is critical - should not continue if update fails
			logrus.Errorf("CRITICAL: Failed to update message status: %v", err)
			logrus.Errorf("Query was: %s", updateQuery)
			logrus.Errorf("Args: %v", args)
			return nil, fmt.Errorf("failed to update message status: %w", err)
		}
		
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected != int64(len(messageIDs)) {
			logrus.Warnf("Expected to update %d rows but updated %d rows", len(messageIDs), rowsAffected)
		}
		logrus.Infof("Successfully updated %d messages with processing_worker_id = %s", rowsAffected, workerID)
	}
