-- CURRENT WORKING SCHEMA REFERENCE
-- This documents the actual schema that should be used
-- DO NOT DROP TABLES - This is just for reference

-- ============================================
-- CORE TABLES
-- ============================================

-- users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP
);

-- user_devices table
CREATE TABLE IF NOT EXISTS user_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_name VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    jid VARCHAR(255),
    status VARCHAR(50) DEFAULT 'offline',
    last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    min_delay_seconds INTEGER DEFAULT 5,
    max_delay_seconds INTEGER DEFAULT 15,
    UNIQUE(user_id, jid)
);

-- user_sessions table
CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- LEADS AND CAMPAIGNS
-- ============================================

-- leads table
CREATE TABLE IF NOT EXISTS leads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id UUID REFERENCES user_devices(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(50) NOT NULL,
    email VARCHAR(255),
    address TEXT,
    status VARCHAR(50) DEFAULT 'prospect',
    niche VARCHAR(100),
    trigger VARCHAR(1000), -- For sequence triggers
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, phone)
);

-- campaigns table  
CREATE TABLE IF NOT EXISTS campaigns (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id UUID REFERENCES user_devices(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    niche VARCHAR(100),
    target_status VARCHAR(50) DEFAULT 'all',
    message TEXT NOT NULL,
    image_url TEXT,
    campaign_date DATE,
    scheduled_date VARCHAR(50),
    time_schedule VARCHAR(10),
    min_delay_seconds INTEGER DEFAULT 5,
    max_delay_seconds INTEGER DEFAULT 15,
    status VARCHAR(50) DEFAULT 'pending',
    ai VARCHAR(10),
    "limit" INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- SEQUENCES
-- ============================================

-- sequences table
CREATE TABLE IF NOT EXISTS sequences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id UUID REFERENCES user_devices(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    niche VARCHAR(100),
    target_status VARCHAR(50) DEFAULT 'all',
    status VARCHAR(50) DEFAULT 'draft',
    trigger VARCHAR(255),
    trigger_prefix VARCHAR(100),
    start_trigger VARCHAR(100),
    end_trigger VARCHAR(100),
    total_days INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 0,
    schedule_time VARCHAR(10),
    min_delay_seconds INTEGER DEFAULT 5,
    max_delay_seconds INTEGER DEFAULT 15,
    contacts_count INTEGER DEFAULT 0,
    total_contacts INTEGER DEFAULT 0,
    active_contacts INTEGER DEFAULT 0,
    completed_contacts INTEGER DEFAULT 0,
    failed_contacts INTEGER DEFAULT 0,
    progress_percentage DECIMAL(5,2) DEFAULT 0,
    last_activity_at TIMESTAMP,
    estimated_completion_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- sequence_steps table
CREATE TABLE IF NOT EXISTS sequence_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sequence_id UUID NOT NULL REFERENCES sequences(id) ON DELETE CASCADE,
    day_number INTEGER NOT NULL,
    trigger VARCHAR(255),
    next_trigger VARCHAR(255),
    trigger_delay_hours INTEGER DEFAULT 24,
    is_entry_point BOOLEAN DEFAULT false,
    message_type VARCHAR(50) DEFAULT 'text',
    message_text TEXT,
    content TEXT,
    media_url TEXT,
    caption TEXT,
    time_schedule VARCHAR(10),
    min_delay_seconds INTEGER DEFAULT 5,
    max_delay_seconds INTEGER DEFAULT 15,
    delay_days INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- sequence_contacts table
CREATE TABLE IF NOT EXISTS sequence_contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sequence_id UUID NOT NULL REFERENCES sequences(id) ON DELETE CASCADE,
    contact_phone VARCHAR(50) NOT NULL,
    contact_name VARCHAR(255),
    current_step INTEGER DEFAULT 0,
    status VARCHAR(50) DEFAULT 'active',
    completed_at TIMESTAMP,
    current_trigger VARCHAR(255),
    next_trigger_time TIMESTAMP,
    processing_device_id UUID,
    last_error TEXT,
    retry_count INTEGER DEFAULT 0,
    assigned_device_id UUID,
    processing_started_at TIMESTAMP,
    sequence_stepid UUID,
    user_id UUID,
    UNIQUE(sequence_id, contact_phone)
);

-- ============================================
-- MESSAGING
-- ============================================

-- broadcast_messages table
CREATE TABLE IF NOT EXISTS broadcast_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES user_devices(id) ON DELETE CASCADE,
    campaign_id INTEGER REFERENCES campaigns(id) ON DELETE SET NULL,
    sequence_id UUID REFERENCES sequences(id) ON DELETE SET NULL,
    recipient_phone VARCHAR(50) NOT NULL,
    message_type VARCHAR(50) NOT NULL,
    content TEXT,
    media_url TEXT,
    status VARCHAR(50) DEFAULT 'pending',
    error_message TEXT,
    scheduled_at TIMESTAMP,
    sent_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    group_id VARCHAR(255),
    group_order INTEGER
);

-- message_analytics table
CREATE TABLE IF NOT EXISTS message_analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES user_devices(id) ON DELETE CASCADE,
    campaign_id INTEGER REFERENCES campaigns(id) ON DELETE SET NULL,
    sequence_id UUID REFERENCES sequences(id) ON DELETE SET NULL,
    recipient_phone VARCHAR(50) NOT NULL,
    message_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,
    read_at TIMESTAMP,
    failed_at TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- WHATSAPP WEB TABLES
-- ============================================

-- whatsapp_chats table
CREATE TABLE IF NOT EXISTS whatsapp_chats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id UUID NOT NULL REFERENCES user_devices(id) ON DELETE CASCADE,
    chat_jid VARCHAR(255) NOT NULL,
    chat_name VARCHAR(255),
    is_group BOOLEAN DEFAULT false,
    is_read_only BOOLEAN DEFAULT false,
    unread_count INTEGER DEFAULT 0,
    last_message_time TIMESTAMP,
    last_message_text TEXT,
    is_archived BOOLEAN DEFAULT false,
    is_pinned BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(device_id, chat_jid)
);

-- whatsapp_messages table
CREATE TABLE IF NOT EXISTS whatsapp_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id UUID NOT NULL REFERENCES user_devices(id) ON DELETE CASCADE,
    chat_jid VARCHAR(255) NOT NULL,
    message_id VARCHAR(255) NOT NULL,
    sender_jid VARCHAR(255) NOT NULL,
    sender_name VARCHAR(255),
    message_type VARCHAR(50) NOT NULL,
    content TEXT,
    media_url TEXT,
    media_mime_type VARCHAR(100),
    media_size BIGINT,
    media_caption TEXT,
    is_from_me BOOLEAN DEFAULT false,
    is_group_msg BOOLEAN DEFAULT false,
    status VARCHAR(50),
    timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(device_id, message_id)
);

-- ============================================
-- TRIGGER PROCESSING TABLES
-- ============================================

-- trigger_process_log table
CREATE TABLE IF NOT EXISTS trigger_process_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sequence_contact_id UUID NOT NULL,
    lead_phone VARCHAR(50) NOT NULL,
    device_id UUID NOT NULL,
    trigger_name VARCHAR(255) NOT NULL,
    status VARCHAR(50),
    error_message TEXT,
    processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- device_load_balance table
CREATE TABLE IF NOT EXISTS device_load_balance (
    device_id UUID PRIMARY KEY,
    messages_hour INTEGER DEFAULT 0,
    messages_today INTEGER DEFAULT 0,
    last_reset_hour TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_reset_day TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_available BOOLEAN DEFAULT true,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- broadcast_locks table
CREATE TABLE IF NOT EXISTS broadcast_locks (
    lock_key VARCHAR(100) PRIMARY KEY,
    locked_by VARCHAR(255),
    locked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP
);

-- ============================================
-- ALL INDEXES
-- ============================================

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_user_devices_user_id ON user_devices(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(token);
CREATE INDEX IF NOT EXISTS idx_leads_device_id ON leads(device_id);
CREATE INDEX IF NOT EXISTS idx_leads_user_id ON leads(user_id);
CREATE INDEX IF NOT EXISTS idx_leads_phone ON leads(phone);
CREATE INDEX IF NOT EXISTS idx_leads_trigger ON leads(trigger) WHERE trigger IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_campaigns_user_id ON campaigns(user_id);
CREATE INDEX IF NOT EXISTS idx_campaigns_date ON campaigns(campaign_date);
CREATE INDEX IF NOT EXISTS idx_sequences_user_id ON sequences(user_id);
CREATE INDEX IF NOT EXISTS idx_sequences_status ON sequences(status);
CREATE INDEX IF NOT EXISTS idx_sequence_steps_sequence_id ON sequence_steps(sequence_id);
CREATE INDEX IF NOT EXISTS idx_sequence_steps_unique_trigger ON sequence_steps(trigger) WHERE trigger IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_sequence_contacts_sequence_id ON sequence_contacts(sequence_id);
CREATE INDEX IF NOT EXISTS idx_sequence_contacts_next_send ON sequence_contacts(next_send_at);
CREATE INDEX IF NOT EXISTS idx_seq_contacts_trigger ON sequence_contacts(current_trigger, next_trigger_time) WHERE status = 'active' AND current_trigger IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_seq_contacts_processing ON sequence_contacts(processing_device_id, processing_started_at) WHERE processing_device_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_seq_contacts_phone ON sequence_contacts(contact_phone);
CREATE INDEX IF NOT EXISTS idx_broadcast_messages_status ON broadcast_messages(status);
CREATE INDEX IF NOT EXISTS idx_message_analytics_user_id ON message_analytics(user_id);
CREATE INDEX IF NOT EXISTS idx_message_analytics_created_at ON message_analytics(created_at);
CREATE INDEX IF NOT EXISTS idx_whatsapp_chats_device_id ON whatsapp_chats(device_id);
CREATE INDEX IF NOT EXISTS idx_whatsapp_chats_updated ON whatsapp_chats(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_whatsapp_messages_device_chat ON whatsapp_messages(device_id, chat_jid);
CREATE INDEX IF NOT EXISTS idx_whatsapp_messages_timestamp ON whatsapp_messages(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_process_log_contact ON trigger_process_log(sequence_contact_id, processed_at DESC);
CREATE INDEX IF NOT EXISTS idx_process_log_device ON trigger_process_log(device_id, processed_at DESC);