-- Fix whatsapp_chats table column issue
-- This script ensures the table has the correct column names

-- 1. First check what columns exist
DO $$ 
DECLARE
    has_name_column boolean;
    has_chat_name_column boolean;
BEGIN
    -- Check if 'name' column exists
    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'whatsapp_chats' AND column_name = 'name'
    ) INTO has_name_column;
    
    -- Check if 'chat_name' column exists
    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'whatsapp_chats' AND column_name = 'chat_name'
    ) INTO has_chat_name_column;
    
    -- If we have 'name' but not 'chat_name', rename it
    IF has_name_column AND NOT has_chat_name_column THEN
        ALTER TABLE whatsapp_chats RENAME COLUMN name TO chat_name;
        RAISE NOTICE 'Renamed column name to chat_name';
    
    -- If we have both (shouldn't happen), drop 'name' column
    ELSIF has_name_column AND has_chat_name_column THEN
        -- First copy any data from name to chat_name if chat_name is empty
        UPDATE whatsapp_chats 
        SET chat_name = name 
        WHERE chat_name IS NULL OR chat_name = '';
        
        -- Then drop the name column
        ALTER TABLE whatsapp_chats DROP COLUMN name;
        RAISE NOTICE 'Dropped duplicate name column';
    
    -- If we have neither, add chat_name
    ELSIF NOT has_name_column AND NOT has_chat_name_column THEN
        ALTER TABLE whatsapp_chats ADD COLUMN chat_name VARCHAR(255) NOT NULL DEFAULT '';
        RAISE NOTICE 'Added chat_name column';
    END IF;
    
    -- Ensure chat_name is NOT NULL
    ALTER TABLE whatsapp_chats ALTER COLUMN chat_name SET NOT NULL;
    
END $$;

-- Show current table structure
\d whatsapp_chats