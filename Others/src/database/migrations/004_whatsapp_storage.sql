-- Create whatsapp_chats table to store chat list
CREATE TABLE IF NOT EXISTS whatsapp_chats (
    id SERIAL PRIMARY KEY,
    device_id VARCHAR(255) NOT NULL,
    chat_jid VARCHAR(255) NOT NULL,
    chat_name VARCHAR(255) NOT NULL,
    is_group BOOLEAN DEFAULT FALSE,
    is_muted BOOLEAN DEFAULT FALSE,
    last_message_text TEXT,
    last_message_time TIMESTAMP,
    unread_count INTEGER DEFAULT 0,
    avatar_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(device_id, chat_jid)
);

-- Create whatsapp_messages table to store message history
CREATE TABLE IF NOT EXISTS whatsapp_messages (
    id SERIAL PRIMARY KEY,
    device_id VARCHAR(255) NOT NULL,
    chat_jid VARCHAR(255) NOT NULL,
    message_id VARCHAR(255) NOT NULL,
    sender_jid VARCHAR(255),
    sender_name VARCHAR(255),
    message_text TEXT,
    message_type VARCHAR(50), -- text, image, video, document, etc
    media_url TEXT,
    is_sent BOOLEAN DEFAULT FALSE,
    is_read BOOLEAN DEFAULT FALSE,
    timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(device_id, message_id)
);

-- Create indexes for performance
CREATE INDEX idx_whatsapp_chats_device_id ON whatsapp_chats(device_id);
CREATE INDEX idx_whatsapp_chats_updated ON whatsapp_chats(updated_at DESC);
CREATE INDEX idx_whatsapp_messages_device_chat ON whatsapp_messages(device_id, chat_jid);
CREATE INDEX idx_whatsapp_messages_timestamp ON whatsapp_messages(timestamp DESC);
