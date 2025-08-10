// Fix the query to use correct status values for sequence statistics
// In GetSequenceDeviceReport function around line 4349

stepStatsQuery := `
    SELECT 
        bm.sequence_stepid,
        COUNT(DISTINCT bm.recipient_phone) as total,
        COUNT(DISTINCT CASE WHEN bm.status = 'sent' AND (bm.error_message IS NULL OR bm.error_message = '') THEN bm.recipient_phone END) as done_send,
        COUNT(DISTINCT CASE WHEN bm.status = 'failed' THEN bm.recipient_phone END) as failed_send,
        COUNT(DISTINCT CASE WHEN bm.status IN ('pending', 'queued') THEN bm.recipient_phone END) as remaining_send
    FROM broadcast_messages bm
    WHERE bm.sequence_id = ? 
    AND bm.device_id = ?
    AND bm.user_id = ?
    AND bm.sequence_stepid IS NOT NULL
    GROUP BY bm.sequence_stepid
`

// Fix the overall query around line 4419
overallQuery := `
    SELECT 
        COUNT(DISTINCT recipient_phone) as total,
        COUNT(DISTINCT CASE WHEN status = 'sent' AND (error_message IS NULL OR error_message = '') THEN recipient_phone END) as done_send,
        COUNT(DISTINCT CASE WHEN status = 'failed' THEN recipient_phone END) as failed_send,
        COUNT(DISTINCT CASE WHEN status IN ('pending', 'queued') THEN recipient_phone END) as remaining_send
    FROM broadcast_messages
    WHERE sequence_id = ? 
    AND user_id = ?
`
