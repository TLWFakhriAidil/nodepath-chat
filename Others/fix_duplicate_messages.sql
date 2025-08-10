-- Fix for duplicate messages being sent
-- This adds worker locking to prevent multiple workers processing same message

-- Step 1: Add columns for worker locking
ALTER TABLE broadcast_messages 
ADD COLUMN processing_worker_id VARCHAR(100) DEFAULT NULL,
ADD COLUMN processing_started_at TIMESTAMP NULL,
ADD INDEX idx_processing_worker (processing_worker_id),
ADD INDEX idx_processing_started (processing_started_at);

-- Step 2: Add cleanup for stuck messages (run this periodically)
-- This resets messages stuck in processing for more than 5 minutes
UPDATE broadcast_messages 
SET processing_worker_id = NULL,
    status = 'pending'
WHERE status = 'processing'
AND processing_started_at < DATE_SUB(NOW(), INTERVAL 5 MINUTE);
