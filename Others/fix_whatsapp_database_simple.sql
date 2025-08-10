-- Simple fix: Just add the missing 'name' column to whatsapp_chats table
-- Run this if you don't want to drop and recreate the table

-- Check if 'name' column exists, if not add it
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'whatsapp_chats' 
        AND column_name = 'name'
    ) THEN
        ALTER TABLE whatsapp_chats ADD COLUMN name TEXT NOT NULL DEFAULT '';
        
        -- Update existing rows to have a name based on chat_jid
        UPDATE whatsapp_chats 
        SET name = SPLIT_PART(chat_jid, '@', 1)
        WHERE name = '';
        
        RAISE NOTICE 'Added name column to whatsapp_chats table';
    ELSE
        RAISE NOTICE 'name column already exists in whatsapp_chats table';
    END IF;
END $$;

-- Fix timestamps in whatsapp_messages
UPDATE whatsapp_messages 
SET timestamp = timestamp / 1000 
WHERE timestamp > 1000000000000;

UPDATE whatsapp_messages 
SET timestamp = EXTRACT(EPOCH FROM NOW())::BIGINT
WHERE timestamp > (EXTRACT(EPOCH FROM NOW()) + 31536000)::BIGINT;

SELECT 'Database fixes applied!' as status;
