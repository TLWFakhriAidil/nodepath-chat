-- SQL Query to check failed messages with sequence_stepid
-- Run this directly on your PostgreSQL database

-- 1. First check if sequence_stepid column exists
SELECT column_name 
FROM information_schema.columns 
WHERE table_name = 'broadcast_messages' 
AND column_name LIKE '%step%';

-- 2. Failed messages with sequence_stepid (adjust column name if different)
SELECT 
    bm.id,
    bm.recipient_phone,
    bm.status,
    bm.error_message,
    bm.created_at,
    bm.scheduled_at,
    bm.sequence_stepid,  -- Change this if column name is different
    bm.sequence_id,
    s.name as sequence_name
FROM broadcast_messages bm
LEFT JOIN sequences s ON bm.sequence_id = s.id
WHERE bm.status = 'failed' 
AND bm.sequence_stepid IS NOT NULL  -- Change this if column name is different
AND bm.error_message IS NOT NULL
ORDER BY bm.created_at DESC
LIMIT 30;

-- 3. Count errors by type for messages with sequence_stepid
SELECT 
    error_message,
    COUNT(*) as count,
    COUNT(DISTINCT sequence_id) as sequences_affected,
    COUNT(DISTINCT sequence_stepid) as steps_affected
FROM broadcast_messages
WHERE status = 'failed'
AND sequence_stepid IS NOT NULL
AND error_message IS NOT NULL
GROUP BY error_message
ORDER BY count DESC;

-- 4. Failures by sequence (with sequence_stepid only)
SELECT 
    s.name as sequence_name,
    s.id as sequence_id,
    COUNT(bm.id) as total_failed,
    COUNT(DISTINCT bm.sequence_stepid) as unique_steps,
    MIN(bm.created_at) as first_failure,
    MAX(bm.created_at) as last_failure
FROM broadcast_messages bm
JOIN sequences s ON bm.sequence_id = s.id
WHERE bm.status = 'failed'
AND bm.sequence_stepid IS NOT NULL
GROUP BY s.id, s.name
ORDER BY total_failed DESC;

-- 5. Failed messages by step details
SELECT 
    ss.day_number,
    ss.trigger,
    ss.message_text,
    COUNT(bm.id) as failures,
    bm.error_message
FROM broadcast_messages bm
JOIN sequence_steps ss ON bm.sequence_stepid = ss.id
WHERE bm.status = 'failed'
AND bm.error_message IS NOT NULL
GROUP BY ss.day_number, ss.trigger, ss.message_text, bm.error_message
ORDER BY failures DESC
LIMIT 20;

-- 6. Summary statistics for messages with sequence_stepid
SELECT 
    COUNT(*) as total_with_stepid,
    COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_with_stepid,
    COUNT(CASE WHEN status = 'sent' THEN 1 END) as sent_with_stepid,
    COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_with_stepid,
    COUNT(CASE WHEN status = 'expired' THEN 1 END) as expired_with_stepid
FROM broadcast_messages
WHERE sequence_stepid IS NOT NULL;

-- 7. Recent failures with sequence_stepid (last 7 days)
SELECT 
    DATE(created_at) as failure_date,
    COUNT(*) as daily_failures,
    COUNT(DISTINCT sequence_id) as sequences_affected
FROM broadcast_messages
WHERE status = 'failed'
AND sequence_stepid IS NOT NULL
AND created_at > NOW() - INTERVAL '7 days'
GROUP BY DATE(created_at)
ORDER BY failure_date DESC;

-- 8. Check if there are any successful messages from the same sequences
SELECT 
    s.name as sequence_name,
    COUNT(CASE WHEN bm.status = 'sent' THEN 1 END) as sent,
    COUNT(CASE WHEN bm.status = 'failed' THEN 1 END) as failed,
    COUNT(CASE WHEN bm.status = 'pending' THEN 1 END) as pending,
    ROUND(COUNT(CASE WHEN bm.status = 'sent' THEN 1 END)::numeric / 
          COUNT(*)::numeric * 100, 2) as success_rate
FROM broadcast_messages bm
JOIN sequences s ON bm.sequence_id = s.id
WHERE bm.sequence_stepid IS NOT NULL
GROUP BY s.id, s.name
ORDER BY failed DESC;
