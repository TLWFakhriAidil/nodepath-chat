-- Debug SQL to check why processing_worker_id is not being set

-- 1. Check if there are any messages with processing_worker_id set
SELECT COUNT(*) as total_with_worker_id
FROM broadcast_messages 
WHERE processing_worker_id IS NOT NULL;

-- 2. Check messages in different statuses
SELECT status, COUNT(*) as count, 
       COUNT(processing_worker_id) as with_worker_id
FROM broadcast_messages
WHERE created_at > DATE_SUB(NOW(), INTERVAL 1 HOUR)
GROUP BY status;

-- 3. Check if messages meet the WHERE conditions for GetPendingMessagesAndLock
SELECT COUNT(*) as eligible_messages
FROM broadcast_messages
WHERE status = 'pending'
AND processing_worker_id IS NULL
AND scheduled_at IS NOT NULL
AND scheduled_at <= DATE_ADD(NOW(), INTERVAL 8 HOUR)
AND scheduled_at >= DATE_ADD(DATE_SUB(NOW(), INTERVAL 10 MINUTE), INTERVAL 8 HOUR);

-- 4. Check actual messages that should be processed
SELECT id, status, processing_worker_id, scheduled_at, 
       NOW() as current_time,
       DATE_ADD(NOW(), INTERVAL 8 HOUR) as max_time,
       DATE_ADD(DATE_SUB(NOW(), INTERVAL 10 MINUTE), INTERVAL 8 HOUR) as min_time
FROM broadcast_messages
WHERE status IN ('pending', 'processing', 'queued')
ORDER BY created_at DESC
LIMIT 10;

-- 5. Test update directly
UPDATE broadcast_messages 
SET processing_worker_id = 'TEST_WORKER_123',
    processing_started_at = NOW()
WHERE id = (SELECT id FROM broadcast_messages WHERE status = 'pending' LIMIT 1);

-- Check if it worked
SELECT id, status, processing_worker_id, processing_started_at
FROM broadcast_messages
WHERE processing_worker_id = 'TEST_WORKER_123';
