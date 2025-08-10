-- Add status column to sequences table
-- This migration adds the status column that was missing from the original schema

-- Add status column if it doesn't exist
ALTER TABLE sequences 
ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'inactive';

-- Update existing sequences to have proper status based on is_active
UPDATE sequences 
SET status = CASE 
    WHEN is_active = true THEN 'active' 
    ELSE 'inactive' 
END;

-- Add index for performance
CREATE INDEX IF NOT EXISTS idx_sequences_status ON sequences(status);

-- Add other missing columns that might be needed
ALTER TABLE sequences 
ADD COLUMN IF NOT EXISTS target_status VARCHAR(50) DEFAULT 'all',
ADD COLUMN IF NOT EXISTS schedule_time VARCHAR(5) DEFAULT '09:00',
ADD COLUMN IF NOT EXISTS min_delay_seconds INTEGER DEFAULT 30,
ADD COLUMN IF NOT EXISTS max_delay_seconds INTEGER DEFAULT 60,
ADD COLUMN IF NOT EXISTS contacts_count INTEGER DEFAULT 0;

-- Add progress tracking columns
ALTER TABLE sequences
ADD COLUMN IF NOT EXISTS total_contacts INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS active_contacts INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS completed_contacts INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS failed_contacts INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS progress_percentage DECIMAL(5,2) DEFAULT 0.00,
ADD COLUMN IF NOT EXISTS last_activity_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS estimated_completion_at TIMESTAMP;

-- Update sequence_contacts table to add missing columns
ALTER TABLE sequence_contacts
ADD COLUMN IF NOT EXISTS current_step INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS enrolled_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN IF NOT EXISTS last_sent_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS next_send_at TIMESTAMP;

-- Update sequence_steps table to add missing columns
ALTER TABLE sequence_steps
ADD COLUMN IF NOT EXISTS day_number INTEGER,
ADD COLUMN IF NOT EXISTS time_schedule VARCHAR(5),
ADD COLUMN IF NOT EXISTS image_url TEXT,
ADD COLUMN IF NOT EXISTS min_delay_seconds INTEGER DEFAULT 30,
ADD COLUMN IF NOT EXISTS max_delay_seconds INTEGER DEFAULT 60;

-- Update sequence_logs table to add missing columns
ALTER TABLE sequence_logs
ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
