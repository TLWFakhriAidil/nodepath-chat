// Fix for MySQL 5.7 - Replace SKIP LOCKED with proper locking

// In broadcast_repository.go, GetPendingMessagesAndLock method:

// Option 1: Remove SKIP LOCKED (will cause blocking but ensures atomic locking)
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
    WHERE bm.device_id = ? 
    AND bm.status = 'pending'
    AND bm.processing_worker_id IS NULL
    AND bm.scheduled_at IS NOT NULL
    AND bm.scheduled_at <= DATE_ADD(NOW(), INTERVAL 8 HOUR)
    AND bm.scheduled_at >= DATE_ADD(DATE_SUB(NOW(), INTERVAL 10 MINUTE), INTERVAL 8 HOUR)
    ORDER BY bm.scheduled_at ASC, bm.group_id, bm.group_order
    LIMIT ?
    FOR UPDATE  -- Remove SKIP LOCKED for MySQL 5.7
`

// Option 2: Better fix - Update status immediately in the SELECT query
// Replace the whole GetPendingMessagesAndLock method with this approach:

func (r *BroadcastRepository) GetPendingMessagesAndLock(deviceID string, limit int) ([]domainBroadcast.BroadcastMessage, error) {
    // Generate unique worker ID
    workerID := fmt.Sprintf("%s_%d_%s", deviceID, time.Now().UnixNano(), uuid.New().String()[:8])
    
    // First, atomically claim messages by updating their status
    result, err := r.db.Exec(`
        UPDATE broadcast_messages bm
        SET bm.status = 'processing',
            bm.processing_worker_id = ?,
            bm.processing_started_at = NOW(),
            bm.updated_at = NOW()
        WHERE bm.device_id = ? 
        AND bm.status = 'pending'
        AND bm.processing_worker_id IS NULL
        AND bm.scheduled_at IS NOT NULL
        AND bm.scheduled_at <= DATE_ADD(NOW(), INTERVAL 8 HOUR)
        AND bm.scheduled_at >= DATE_ADD(DATE_SUB(NOW(), INTERVAL 10 MINUTE), INTERVAL 8 HOUR)
        ORDER BY bm.scheduled_at ASC, bm.group_id, bm.group_order
        LIMIT ?
    `, workerID, deviceID, limit)
    
    if err != nil {
        return nil, fmt.Errorf("failed to claim messages: %w", err)
    }
    
    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return []domainBroadcast.BroadcastMessage{}, nil
    }
    
    // Now fetch the messages we just claimed
    query := `
        SELECT bm.id, bm.user_id, bm.device_id, bm.campaign_id, bm.sequence_id, 
            bm.recipient_phone, bm.recipient_name, bm.message_type, bm.content AS message, bm.media_url, 
            bm.scheduled_at, bm.group_id, bm.group_order, bm.sequence_stepid,
            COALESCE(c.min_delay_seconds, ss.min_delay_seconds, s.min_delay_seconds, 10) AS min_delay,
            COALESCE(c.max_delay_seconds, ss.max_delay_seconds, s.max_delay_seconds, 30) AS max_delay
        FROM broadcast_messages bm
        LEFT JOIN campaigns c ON bm.campaign_id = c.id
        LEFT JOIN sequences s ON bm.sequence_id = s.id
        LEFT JOIN sequence_steps ss ON bm.sequence_stepid = ss.id
        WHERE bm.processing_worker_id = ?
        ORDER BY bm.scheduled_at ASC, bm.group_id, bm.group_order
    `
    
    rows, err := r.db.Query(query, workerID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    // ... rest of the method to scan rows into messages ...
}