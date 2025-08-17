-- Create table to store recent WhatsApp messages
-- Limited to 20 messages per chat (like WhatsApp Web)
CREATE TABLE IF NOT EXISTS whatsapp_messages (
    device_id TEXT NOT NULL,
    chat_jid TEXT NOT NULL,
    message_id TEXT NOT NULL,
    sender_jid TEXT,
    message_text TEXT,
    message_type TEXT DEFAULT 'text',
    timestamp BIGINT,
    is_from_me BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (device_id, chat_jid, message_id)
);

-- Index for fast chat queries
CREATE INDEX IF NOT EXISTS idx_whatsapp_messages_chat ON whatsapp_messages(device_id, chat_jid, timestamp DESC);

-- Keep only recent 20 messages per chat (like WhatsApp Web)
CREATE OR REPLACE FUNCTION limit_chat_messages() 
RETURNS TRIGGER AS $$
BEGIN
    -- Delete old messages if more than 20 in this chat
    DELETE FROM whatsapp_messages 
    WHERE device_id = NEW.device_id 
    AND chat_jid = NEW.chat_jid
    AND message_id NOT IN (
        SELECT message_id 
        FROM whatsapp_messages 
        WHERE device_id = NEW.device_id 
        AND chat_jid = NEW.chat_jid
        ORDER BY timestamp DESC 
        LIMIT 20
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply trigger
DROP TRIGGER IF EXISTS limit_messages_trigger ON whatsapp_messages;
CREATE TRIGGER limit_messages_trigger 
AFTER INSERT ON whatsapp_messages 
FOR EACH ROW EXECUTE FUNCTION limit_chat_messages();