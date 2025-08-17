-- Add delay settings to devices
ALTER TABLE user_devices ADD COLUMN IF NOT EXISTS min_delay_seconds INTEGER DEFAULT 5;
ALTER TABLE user_devices ADD COLUMN IF NOT EXISTS max_delay_seconds INTEGER DEFAULT 15;

-- Add queue table for message processing
CREATE TABLE IF NOT EXISTS message_queue (
    id VARCHAR(255) PRIMARY KEY,
    device_id VARCHAR(255) NOT NULL,
    message_type VARCHAR(50) NOT NULL, -- campaign, sequence
    reference_id VARCHAR(255) NOT NULL, -- campaign_id or sequence_id
    contact_phone VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    media_url TEXT,
    caption TEXT,
    priority INTEGER DEFAULT 0,
    status VARCHAR(50) DEFAULT 'pending', -- pending, processing, sent, failed
    scheduled_at TIMESTAMP,
    processed_at TIMESTAMP,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (device_id) REFERENCES user_devices(id) ON DELETE CASCADE,
    INDEX idx_queue_status (status),
    INDEX idx_queue_device (device_id),
    INDEX idx_queue_scheduled (scheduled_at)
);

-- Add broadcast job tracking
CREATE TABLE IF NOT EXISTS broadcast_jobs (
    id VARCHAR(255) PRIMARY KEY,
    job_type VARCHAR(50) NOT NULL, -- campaign, sequence
    reference_id VARCHAR(255) NOT NULL,
    device_id VARCHAR(255) NOT NULL,
    total_contacts INTEGER NOT NULL,
    processed_contacts INTEGER DEFAULT 0,
    successful_contacts INTEGER DEFAULT 0,
    failed_contacts INTEGER DEFAULT 0,
    status VARCHAR(50) DEFAULT 'running', -- running, completed, failed
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    FOREIGN KEY (device_id) REFERENCES user_devices(id) ON DELETE CASCADE
);