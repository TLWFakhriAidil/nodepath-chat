-- Fix column name mismatch for sequence_contacts table
-- This ensures both column names exist for compatibility

-- Add next_send_at column if it doesn't exist (expected by the model)
ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS next_send_at TIMESTAMP;

-- Add next_trigger_time column if it doesn't exist (used by migrations)
ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS next_trigger_time TIMESTAMP;

-- Sync data between columns if one has data and the other doesn't
UPDATE sequence_contacts 
SET next_send_at = next_trigger_time 
WHERE next_send_at IS NULL AND next_trigger_time IS NOT NULL;

UPDATE sequence_contacts 
SET next_trigger_time = next_send_at 
WHERE next_trigger_time IS NULL AND next_send_at IS NOT NULL;

-- Add trigger to keep both columns in sync going forward
CREATE OR REPLACE FUNCTION sync_sequence_contact_times()
RETURNS TRIGGER AS $$
BEGIN
    -- If next_send_at is updated, update next_trigger_time
    IF NEW.next_send_at IS DISTINCT FROM OLD.next_send_at THEN
        NEW.next_trigger_time = NEW.next_send_at;
    END IF;
    
    -- If next_trigger_time is updated, update next_send_at
    IF NEW.next_trigger_time IS DISTINCT FROM OLD.next_trigger_time THEN
        NEW.next_send_at = NEW.next_trigger_time;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop trigger if exists and recreate
DROP TRIGGER IF EXISTS sync_sequence_times_trigger ON sequence_contacts;

CREATE TRIGGER sync_sequence_times_trigger
BEFORE UPDATE ON sequence_contacts
FOR EACH ROW
EXECUTE FUNCTION sync_sequence_contact_times();

-- Ensure all other critical columns exist
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS ai VARCHAR(10);
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS "limit" INTEGER DEFAULT 0;
ALTER TABLE leads ADD COLUMN IF NOT EXISTS trigger VARCHAR(1000);
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS trigger VARCHAR(255);
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS next_trigger VARCHAR(255);
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS trigger_delay_hours INTEGER DEFAULT 24;
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS is_entry_point BOOLEAN DEFAULT false;
ALTER TABLE user_devices ADD COLUMN IF NOT EXISTS min_delay_seconds INTEGER DEFAULT 5;
ALTER TABLE user_devices ADD COLUMN IF NOT EXISTS max_delay_seconds INTEGER DEFAULT 15;

-- Create missing indexes if they don't exist
CREATE INDEX IF NOT EXISTS idx_sequence_contacts_next_send ON sequence_contacts(next_send_at);
CREATE INDEX IF NOT EXISTS idx_seq_contacts_trigger ON sequence_contacts(current_trigger, next_trigger_time) 
WHERE status = 'active' AND current_trigger IS NOT NULL;