-- Fix timestamp issues in whatsapp_messages table
-- Run this SQL to fix the timestamp out of range errors

-- 1. Fix timestamps that are in milliseconds (too large)
UPDATE whatsapp_messages 
SET timestamp = timestamp / 1000 
WHERE timestamp > 1000000000000;

-- 2. Fix timestamps that are in the future (more than 1 year from now)
UPDATE whatsapp_messages 
SET timestamp = EXTRACT(EPOCH FROM NOW())::BIGINT
WHERE timestamp > (EXTRACT(EPOCH FROM NOW()) + 31536000)::BIGINT;

-- 3. Show how many messages were fixed
SELECT 
    COUNT(*) as total_messages,
    COUNT(CASE WHEN timestamp > 1000000000000 THEN 1 END) as millisecond_timestamps,
    COUNT(CASE WHEN timestamp > (EXTRACT(EPOCH FROM NOW()) + 31536000)::BIGINT THEN 1 END) as future_timestamps
FROM whatsapp_messages;

-- 4. Create a trigger to automatically fix future timestamps
CREATE OR REPLACE FUNCTION validate_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    -- If timestamp is too large (likely milliseconds or corrupted)
    IF NEW.timestamp > (EXTRACT(EPOCH FROM NOW()) + 31536000) THEN
        -- If it's likely milliseconds
        IF NEW.timestamp > 1000000000000 THEN
            NEW.timestamp = NEW.timestamp / 1000;
        ELSE
            -- Use current timestamp as fallback
            NEW.timestamp = EXTRACT(EPOCH FROM NOW());
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop existing trigger if exists
DROP TRIGGER IF EXISTS validate_whatsapp_message_timestamp ON whatsapp_messages;

-- Create trigger to validate future timestamps
CREATE TRIGGER validate_whatsapp_message_timestamp
BEFORE INSERT OR UPDATE ON whatsapp_messages
FOR EACH ROW
EXECUTE FUNCTION validate_timestamp();

SELECT 'Timestamp fixes applied!' as status;
