SELECT 
    COUNT(*) as total,
    status,
    COUNT(processing_worker_id) as with_worker_id,
    DATE(created_at) as date
FROM broadcast_messages
WHERE created_at > DATE_SUB(NOW(), INTERVAL 7 DAY)
GROUP BY status, DATE(created_at)
ORDER BY date DESC, status;

-- Check scheduled_at values
SELECT 
    COUNT(*) as count,
    CASE 
        WHEN scheduled_at IS NULL THEN 'NULL'
        WHEN scheduled_at <= NOW() THEN 'PAST'
        WHEN scheduled_at > NOW() AND scheduled_at <= DATE_ADD(NOW(), INTERVAL 1 HOUR) THEN 'NEXT_HOUR'
        WHEN scheduled_at > DATE_ADD(NOW(), INTERVAL 1 HOUR) AND scheduled_at <= DATE_ADD(NOW(), INTERVAL 8 HOUR) THEN 'NEXT_8_HOURS'
        ELSE 'FUTURE'
    END as schedule_status
FROM broadcast_messages
WHERE status = 'pending'
GROUP BY schedule_status;

-- Check the exact time comparison
SELECT 
    id,
    status,
    scheduled_at,
    NOW() as current_time,
    DATE_ADD(NOW(), INTERVAL 8 HOUR) as max_time,
    DATE_ADD(DATE_SUB(NOW(), INTERVAL 10 MINUTE), INTERVAL 8 HOUR) as min_time,
    CASE
        WHEN scheduled_at <= DATE_ADD(NOW(), INTERVAL 8 HOUR) 
         AND scheduled_at >= DATE_ADD(DATE_SUB(NOW(), INTERVAL 10 MINUTE), INTERVAL 8 HOUR) 
        THEN 'ELIGIBLE'
        ELSE 'NOT_ELIGIBLE'
    END as eligibility
FROM broadcast_messages
WHERE status = 'pending'
AND processing_worker_id IS NULL
ORDER BY scheduled_at DESC
LIMIT 10;
