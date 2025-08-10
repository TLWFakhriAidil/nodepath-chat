-- Add trigger column to leads (comma-separated for multiple sequences)
ALTER TABLE leads ADD COLUMN IF NOT EXISTS trigger VARCHAR(1000);

-- Create index for trigger searches
CREATE INDEX IF NOT EXISTS idx_leads_trigger ON leads(trigger) WHERE trigger IS NOT NULL;

-- Add fields for better trigger-based processing
ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS current_trigger VARCHAR(255);
ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS next_trigger_time TIMESTAMP;
ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS processing_device_id UUID;
ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS processing_started_at TIMESTAMP;
ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS last_error TEXT;
ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS retry_count INT DEFAULT 0;

-- Create indexes for fast processing
CREATE INDEX IF NOT EXISTS idx_seq_contacts_trigger ON sequence_contacts(current_trigger, next_trigger_time) 
WHERE status = 'active' AND current_trigger IS NOT NULL;

-- Ensure sequence_steps has proper trigger flow
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS next_trigger VARCHAR(255);
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS trigger_delay_hours INT DEFAULT 24;
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS is_entry_point BOOLEAN DEFAULT false;

-- Create unique constraint on trigger
CREATE UNIQUE INDEX IF NOT EXISTS idx_sequence_steps_unique_trigger ON sequence_steps(trigger) WHERE trigger IS NOT NULL;

-- Mark entry points for each sequence
UPDATE sequence_steps SET is_entry_point = true WHERE day_number = 1;

-- Create device load balance table
CREATE TABLE IF NOT EXISTS device_load_balance (
    device_id UUID PRIMARY KEY,
    messages_hour INT DEFAULT 0,
    messages_today INT DEFAULT 0,
    last_reset_hour TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_reset_day TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_available BOOLEAN DEFAULT true,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create unique constraint to prevent duplicate enrollments
ALTER TABLE sequence_contacts ADD CONSTRAINT IF NOT EXISTS uq_sequence_contact 
UNIQUE (sequence_id, contact_phone);