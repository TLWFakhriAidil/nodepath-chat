-- Add recipient_name column to broadcast_messages table
ALTER TABLE broadcast_messages ADD COLUMN IF NOT EXISTS recipient_name VARCHAR(255);

-- Update existing records to use phone number as name if null
UPDATE broadcast_messages SET recipient_name = recipient_phone WHERE recipient_name IS NULL;
