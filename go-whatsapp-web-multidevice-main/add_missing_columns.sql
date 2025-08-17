-- Add missing columns that the application expects
-- Run this on your database to match what the code expects

-- Add trigger column to leads table if it doesn't exist
ALTER TABLE leads ADD COLUMN IF NOT EXISTS trigger VARCHAR(1000);

-- Add priority column to sequences table if it doesn't exist
ALTER TABLE sequences ADD COLUMN IF NOT EXISTS priority INTEGER DEFAULT 0;

-- Add missing columns to sequence_contacts for trigger-based processing
ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS current_trigger VARCHAR(255);
ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS processing_device_id UUID;
ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS processing_started_at TIMESTAMP;
ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS last_error TEXT;
ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS retry_count INTEGER DEFAULT 0;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_leads_trigger ON leads(trigger) WHERE trigger IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_sequences_priority ON sequences(priority);
CREATE INDEX IF NOT EXISTS idx_seq_contacts_processing ON sequence_contacts(processing_device_id, processing_started_at) WHERE processing_device_id IS NOT NULL;

-- Note: If you get errors about columns already existing, that's fine - it means they're already there