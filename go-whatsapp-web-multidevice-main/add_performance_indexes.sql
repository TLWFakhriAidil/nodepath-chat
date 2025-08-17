-- Performance Optimization for Broadcast Messages
-- Date: August 12, 2025
-- Purpose: Speed up message pickup queries

-- This index optimizes the main query used in GetPendingMessagesAndLock()
-- It covers: status, device_id, and scheduled_at columns
-- This makes the WHERE clause and ORDER BY much faster

-- Check if index exists first
SELECT COUNT(*) AS index_exists
FROM information_schema.statistics 
WHERE table_schema = DATABASE()
AND table_name = 'broadcast_messages' 
AND index_name = 'idx_broadcast_optimize';

-- Create the index if it doesn't exist
-- This will speed up queries like:
-- SELECT * FROM broadcast_messages 
-- WHERE status = 'pending' AND device_id = ? AND scheduled_at <= ?
-- ORDER BY scheduled_at ASC
CREATE INDEX IF NOT EXISTS idx_broadcast_optimize 
ON broadcast_messages(status, device_id, scheduled_at);

-- Additional helpful indexes for common queries
-- For finding devices with pending messages
CREATE INDEX IF NOT EXISTS idx_status_scheduled 
ON broadcast_messages(status, scheduled_at);

-- For campaign/sequence specific queries
CREATE INDEX IF NOT EXISTS idx_campaign_status 
ON broadcast_messages(campaign_id, status);

CREATE INDEX IF NOT EXISTS idx_sequence_status 
ON broadcast_messages(sequence_id, status);

-- Show all indexes to verify
SHOW INDEX FROM broadcast_messages;
