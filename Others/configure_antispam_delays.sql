-- Configure Anti-Spam Delays for Sequences
-- Run this SQL to enable anti-spam features on your sequences

-- 1. Set default delays for all sequences that don't have them
UPDATE sequences 
SET min_delay_seconds = COALESCE(min_delay_seconds, 5),
    max_delay_seconds = COALESCE(max_delay_seconds, 15)
WHERE min_delay_seconds IS NULL OR max_delay_seconds IS NULL;

-- 2. Set specific delays for active sequences (adjust as needed)
UPDATE sequences 
SET min_delay_seconds = 10,
    max_delay_seconds = 30
WHERE status = 'active' AND min_delay_seconds < 10;

-- 3. Set device fallback delays
UPDATE user_devices 
SET min_delay_seconds = COALESCE(min_delay_seconds, 5),
    max_delay_seconds = COALESCE(max_delay_seconds, 15)
WHERE status IN ('online', 'connected') 
  AND (min_delay_seconds IS NULL OR max_delay_seconds IS NULL);

-- 4. View current delay settings
SELECT 
    name,
    status,
    min_delay_seconds,
    max_delay_seconds,
    CASE 
        WHEN min_delay_seconds IS NULL THEN 'No delays set'
        ELSE min_delay_seconds || '-' || max_delay_seconds || ' seconds'
    END as delay_range
FROM sequences
ORDER BY status, name;

-- 5. Check which devices have delays configured
SELECT 
    device_name,
    status,
    min_delay_seconds,
    max_delay_seconds
FROM user_devices
WHERE status IN ('online', 'connected')
ORDER BY device_name;

-- 6. Add missing columns for sequence processing (if needed)
DO $$
BEGIN
    -- Add next_trigger_time for scheduling
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'sequence_contacts' 
                   AND column_name = 'next_trigger_time') THEN
        ALTER TABLE sequence_contacts ADD COLUMN next_trigger_time TIMESTAMP;
        COMMENT ON COLUMN sequence_contacts.next_trigger_time IS 'When to process next message';
    END IF;

    -- Add current_trigger for tracking position
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'sequence_contacts' 
                   AND column_name = 'current_trigger') THEN
        ALTER TABLE sequence_contacts ADD COLUMN current_trigger VARCHAR(255);
        COMMENT ON COLUMN sequence_contacts.current_trigger IS 'Current trigger being processed';
    END IF;

    -- Add processing_device_id to prevent double processing
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'sequence_contacts' 
                   AND column_name = 'processing_device_id') THEN
        ALTER TABLE sequence_contacts ADD COLUMN processing_device_id UUID;
        COMMENT ON COLUMN sequence_contacts.processing_device_id IS 'Device currently processing this contact';
    END IF;
END$$;

-- 7. Create index for better performance
CREATE INDEX IF NOT EXISTS idx_seq_contacts_processing 
ON sequence_contacts(status, next_trigger_time) 
WHERE status = 'active' AND processing_device_id IS NULL;

-- Done! Your sequences now have anti-spam protection
