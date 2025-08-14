-- Sequence tables migration
-- Run this to add sequence functionality

-- Main sequences table
CREATE TABLE IF NOT EXISTS sequences (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    device_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    niche VARCHAR(255), -- Auto-trigger based on lead niche
    total_days INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (device_id) REFERENCES user_devices(id) ON DELETE CASCADE
);

-- Sequence steps (messages for each day)
CREATE TABLE IF NOT EXISTS sequence_steps (
    id VARCHAR(255) PRIMARY KEY,
    sequence_id VARCHAR(255) NOT NULL,
    day INTEGER NOT NULL,
    message_type VARCHAR(50) NOT NULL, -- text, image, video, document
    content TEXT NOT NULL,
    media_url TEXT,
    caption TEXT,
    send_time VARCHAR(5) NOT NULL, -- HH:MM format
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sequence_id) REFERENCES sequences(id) ON DELETE CASCADE,
    UNIQUE(sequence_id, day)
);

-- Contacts enrolled in sequences
CREATE TABLE IF NOT EXISTS sequence_contacts (
    id VARCHAR(255) PRIMARY KEY,
    sequence_id VARCHAR(255) NOT NULL,
    contact_phone VARCHAR(255) NOT NULL,
    contact_name VARCHAR(255),
    current_day INTEGER DEFAULT 0,
    status VARCHAR(50) DEFAULT 'active', -- active, completed, paused
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_message_at TIMESTAMP,
    completed_at TIMESTAMP,
    FOREIGN KEY (sequence_id) REFERENCES sequences(id) ON DELETE CASCADE,
    UNIQUE(sequence_id, contact_phone)
);

-- Log of messages sent
CREATE TABLE IF NOT EXISTS sequence_logs (
    id VARCHAR(255) PRIMARY KEY,
    sequence_id VARCHAR(255) NOT NULL,
    contact_id VARCHAR(255) NOT NULL,
    step_id VARCHAR(255) NOT NULL,
    day INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL, -- sent, delivered, read, failed
    message_id VARCHAR(255),
    error_message TEXT,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sequence_id) REFERENCES sequences(id) ON DELETE CASCADE,
    FOREIGN KEY (contact_id) REFERENCES sequence_contacts(id) ON DELETE CASCADE,
    FOREIGN KEY (step_id) REFERENCES sequence_steps(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX idx_sequences_user_id ON sequences(user_id);
CREATE INDEX idx_sequences_device_id ON sequences(device_id);
CREATE INDEX idx_sequence_contacts_status ON sequence_contacts(status);
CREATE INDEX idx_sequence_contacts_current_day ON sequence_contacts(current_day);
CREATE INDEX idx_sequence_logs_sent_at ON sequence_logs(sent_at);