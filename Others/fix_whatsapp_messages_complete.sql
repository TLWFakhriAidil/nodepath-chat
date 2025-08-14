-- Complete fix for whatsapp_messages table schema mismatch
-- This fixes the "column message_secrets does not exist" error

-- Step 1: Add message_secrets column if it doesn't exist
ALTER TABLE whatsapp_messages ADD COLUMN IF NOT EXISTS message_secrets TEXT;

-- Step 2: Copy data from media_url to message_secrets if media_url exists
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name = 'whatsapp_messages' AND column_name = 'media_url') THEN
        UPDATE whatsapp_messages 
        SET message_secrets = media_url 
        WHERE message_secrets IS NULL AND media_url IS NOT NULL;
    END IF;
END $$;

-- Step 3: Add missing columns that might be needed
ALTER TABLE whatsapp_messages ADD COLUMN IF NOT EXISTS sender_name VARCHAR(255);
ALTER TABLE whatsapp_messages ADD COLUMN IF NOT EXISTS media_url TEXT;
ALTER TABLE whatsapp_messages ADD COLUMN IF NOT EXISTS is_sent BOOLEAN DEFAULT FALSE;
ALTER TABLE whatsapp_messages ADD COLUMN IF NOT EXISTS is_read BOOLEAN DEFAULT FALSE;

-- Step 4: Ensure proper indexes exist
CREATE INDEX IF NOT EXISTS idx_whatsapp_messages_device_chat ON whatsapp_messages(device_id, chat_jid);
CREATE INDEX IF NOT EXISTS idx_whatsapp_messages_timestamp ON whatsapp_messages(timestamp DESC);

-- Step 5: Verify the fix
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'whatsapp_messages' 