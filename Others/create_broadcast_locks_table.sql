-- Create broadcast locks table to prevent overlapping campaigns/sequences
CREATE TABLE IF NOT EXISTS broadcast_locks (
    user_id VARCHAR(255) PRIMARY KEY,
    broadcast_type VARCHAR(50) NOT NULL, -- 'campaign' or 'sequence'
    broadcast_id VARCHAR(255) NOT NULL,
    locked_at TIMESTAMP DEFAULT NOW(),
    expected_finish TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Add indexes
CREATE INDEX IF NOT EXISTS idx_broadcast_locks_locked_at ON broadcast_locks(locked_at);
CREATE INDEX IF NOT EXISTS idx_broadcast_locks_type ON broadcast_locks(broadcast_type);

-- Add error_message column to campaigns for conflict messages
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS error_message TEXT;

-- Clean up old locks on startup
DELETE FROM broadcast_locks WHERE locked_at < NOW() - INTERVAL '4 hours';