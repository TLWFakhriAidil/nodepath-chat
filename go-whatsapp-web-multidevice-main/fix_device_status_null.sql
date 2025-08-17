-- Fix NULL device_status in broadcast_messages table
UPDATE broadcast_messages 
SET device_status = 'unknown' 
WHERE device_status IS NULL;

-- Make device_status NOT NULL with default value
ALTER TABLE broadcast_messages 
MODIFY COLUMN device_status VARCHAR(50) NOT NULL DEFAULT 'unknown';

-- Show count of messages with issues
SELECT 
    COUNT(*) as total_messages,
    SUM(CASE WHEN device_status IS NULL THEN 1 ELSE 0 END) as null_status_count,
    SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending_count
FROM broadcast_messages;
