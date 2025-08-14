-- Fix missing "key" column in whatsmeow_message_secrets table
-- This adds the required column if it doesn't exist

-- Check and add the "key" column to whatsmeow_message_secrets
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'whatsmeow_message_secrets' 
        AND column_name = 'key'
    ) THEN
        ALTER TABLE whatsmeow_message_secrets 
        ADD COLUMN key bytea;
        
        RAISE NOTICE 'Added "key" column to whatsmeow_message_secrets table';
    ELSE
        RAISE NOTICE '"key" column already exists in whatsmeow_message_secrets table';
    END IF;
END $$;

-- Also ensure the table has the correct structure
-- This creates the table if it doesn't exist with all required columns
CREATE TABLE IF NOT EXISTS whatsmeow_message_secrets (
    our_jid text,
    chat_jid text,
    sender_jid text,
    message_id text,
    key bytea,
    PRIMARY KEY (our_jid, chat_jid, sender_jid, message_id)
);

-- Add foreign key if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.table_constraints
        WHERE constraint_name = 'whatsmeow_message_secrets_our_jid_fkey'
    ) THEN
        ALTER TABLE whatsmeow_message_secrets
        ADD CONSTRAINT whatsmeow_message_secrets_our_jid_fkey
        FOREIGN KEY (our_jid) REFERENCES whatsmeow_device(jid) ON DELETE CASCADE;
    END IF;
END $$;

-- Verify the structure
SELECT column_name, data_type, is_nullable 
FROM information_schema.columns 
WHERE table_name = 'whatsmeow_message_secrets'
ORDER BY ordinal_position;