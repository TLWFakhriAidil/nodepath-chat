-- Fix WhatsApp Web Database Tables
-- Run this SQL script in your PostgreSQL database to fix the tables

-- 1. First, check if whatsapp_chats table exists and has wrong schema
SELECT 'Checking whatsapp_chats table...' as status;

-- Drop the old table if it exists (with wrong schema)
DROP TABLE IF EXISTS whatsapp_chats CASCADE;

-- Create whatsapp_chats table with correct schema
CREATE TABLE whatsapp_chats (
    id SERIAL PRIMARY KEY,
    device_id TEXT NOT NULL,
    chat_jid TEXT NOT NULL,
    name TEXT NOT NULL,
    last_message_time TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(device_id, chat_jid)
);

-- Create index for performance
CREATE INDEX idx_chats_device_time ON whatsapp_chats(device_id, last_message_time DESC);

SELECT 'whatsapp_chats table created successfully!' as status;

-- 2. Fix whatsapp_messages table timestamps
SELECT 'Fixing whatsapp_messages timestamps...' as status;

-- First, update any timestamps that are too large (likely milliseconds)
UPDATE whatsapp_messages 
SET timestamp = timestamp / 1000 
WHERE timestamp > 1000000000000;

-- Update any timestamps that are still in the future (more than 1 year from now)
UPDATE whatsapp_messages 
SET timestamp = EXTRACT(EPOCH FROM NOW())::BIGINT
WHERE timestamp > (EXTRACT(EPOCH FROM NOW()) + 31536000)::BIGINT;

-- Create function to validate timestamps
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

SELECT 'whatsapp_messages timestamps fixed!' as status;

-- 3. Show current status
SELECT 'Current database status:' as status;

-- Count chats
SELECT COUNT(*) as total_chats FROM whatsapp_chats;

-- Count messages
SELECT COUNT(*) as total_messages FROM whatsapp_messages;

-- Show any messages with invalid timestamps
SELECT COUNT(*) as future_timestamps 
FROM whatsapp_messages 
WHERE timestamp > (EXTRACT(EPOCH FROM NOW()) + 31536000)::BIGINT;

SELECT 'Database fix completed!' as status;
