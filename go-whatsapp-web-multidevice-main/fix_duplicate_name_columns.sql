-- Fix duplicate columns in whatsapp_chats table
-- You have both 'name' and 'chat_name' columns, which is causing issues

-- 1. First, copy any data from 'name' to 'chat_name' if chat_name is empty
UPDATE whatsapp_chats 
SET chat_name = name 
WHERE (chat_name IS NULL OR chat_name = '') 
  AND name IS NOT NULL 
  AND name != '';

-- 2. Now drop the old 'name' column
ALTER TABLE whatsapp_chats DROP COLUMN IF EXISTS name;

-- 3. Ensure chat_name has proper constraints
ALTER TABLE whatsapp_chats ALTER COLUMN chat_name SET NOT NULL;

-- 4. Show the updated table structure
\d whatsapp_chats