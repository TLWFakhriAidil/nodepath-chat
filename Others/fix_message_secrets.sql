-- Quick fix to add message_secrets column to whatsapp_messages table
ALTER TABLE whatsapp_messages ADD COLUMN IF NOT EXISTS message_secrets TEXT;

-- Copy data from media_url to message_secrets if media_url exists
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name = 'whatsapp_messages' AND column_name = 'media_url') THEN
        UPDATE whatsapp_messages SET message_secrets = media_url WHERE message_secrets IS NULL;
    END IF;
END $$;
