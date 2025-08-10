-- Fix scheduled_time to use VARCHAR for simplicity
-- This will store time in HH:MM format

-- 1. Add new column as VARCHAR
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS scheduled_time_str VARCHAR(10);

-- 2. Copy existing time data to new column
UPDATE campaigns 
SET scheduled_time_str = TO_CHAR(scheduled_time, 'HH24:MI')
WHERE scheduled_time IS NOT NULL;

-- 3. Drop old column and rename new one
ALTER TABLE campaigns DROP COLUMN IF EXISTS scheduled_time;
ALTER TABLE campaigns RENAME COLUMN scheduled_time_str TO scheduled_time;

-- 4. Update sequences table schedule_time to ensure it's VARCHAR
ALTER TABLE sequences ALTER COLUMN schedule_time TYPE VARCHAR(10);

-- 5. Update sequence_steps table schedule_time to ensure it's VARCHAR  
ALTER TABLE sequence_steps ALTER COLUMN schedule_time TYPE VARCHAR(10);

-- 6. Create indexes for better performance with worker queries
CREATE INDEX IF NOT EXISTS idx_campaigns_status_date ON campaigns(status, campaign_date);
CREATE INDEX IF NOT EXISTS idx_campaigns_user_status ON campaigns(user_id, status);
CREATE INDEX IF NOT EXISTS idx_sequences_status ON sequences(status);
CREATE INDEX IF NOT EXISTS idx_broadcast_messages_status ON broadcast_messages(status, scheduled_at);

-- 7. Add worker tracking table for monitoring
CREATE TABLE IF NOT EXISTS worker_status (
    id SERIAL PRIMARY KEY,
    device_id UUID NOT NULL,
    worker_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) DEFAULT 'idle',
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    messages_processed INTEGER DEFAULT 0,
    messages_failed INTEGER DEFAULT 0,
    current_queue_size INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(device_id, worker_type)
);

-- 8. Create index for worker status queries
CREATE INDEX IF NOT EXISTS idx_worker_status_device ON worker_status(device_id);
CREATE INDEX IF NOT EXISTS idx_worker_status_updated ON worker_status(updated_at);
