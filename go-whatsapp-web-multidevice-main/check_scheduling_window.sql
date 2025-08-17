-- Check the scheduled_at distribution
SELECT 
    DATE(scheduled_at) as scheduled_date,
    COUNT(*) as count,
    MIN(scheduled_at) as earliest,
    MAX(scheduled_at) as latest,
    NOW() as current_db_time,
    DATE_ADD(NOW(), INTERVAL 8 HOUR) as max_eligible_time
FROM broadcast_messages
WHERE status = 'pending'
GROUP BY DATE(scheduled_at)
ORDER BY scheduled_date;

-- Check messages that should be sent NOW (within the 8-hour window)
SELECT 
    COUNT(*) as eligible_count,
    MIN(scheduled_at) as next_scheduled,
    TIMESTAMPDIFF(HOUR, NOW(), MIN(scheduled_at)) as hours_until_next
FROM broadcast_messages
WHERE status = 'pending'
AND processing_worker_id IS NULL
AND scheduled_at IS NOT NULL
AND scheduled_at <= DATE_ADD(NOW(), INTERVAL 8 HOUR)
AND scheduled_at >= DATE_ADD(DATE_SUB(NOW(), INTERVAL 10 MINUTE), INTERVAL 8 HOUR);

-- Messages that would be eligible if we remove the 8-hour window
SELECT 
    COUNT(*) as would_be_eligible,
    MIN(scheduled_at) as next_scheduled
FROM broadcast_messages
WHERE status = 'pending'
AND processing_worker_id IS NULL
AND scheduled_at IS NOT NULL
AND scheduled_at <= NOW();
