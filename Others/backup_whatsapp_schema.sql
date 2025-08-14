-- ================================================
-- Complete WhatsApp Database Backup Script
-- Generated: Current Date
-- ================================================

-- This script contains the complete schema for all WhatsApp tables
-- Use this to restore if tables are dropped

-- 1. WhatsApp Device Table (Main table)
CREATE TABLE IF NOT EXISTS whatsmeow_device (
    jid TEXT PRIMARY KEY,
    lid TEXT,
    registration_id BIGINT NOT NULL,
    noise_key BYTEA NOT NULL,
    identity_key BYTEA NOT NULL,
    signed_pre_key BYTEA NOT NULL,
    signed_pre_key_id INTEGER NOT NULL,
    signed_pre_key_sig BYTEA NOT NULL,
    adv_key BYTEA,
    adv_details BYTEA,
    adv_account_sig BYTEA,
    adv_account_sig_key BYTEA,
    adv_device_sig BYTEA,
    platform TEXT DEFAULT '',
    business_name TEXT DEFAULT '',
    push_name TEXT DEFAULT '',
    facebook_uuid TEXT,
    initialized BOOLEAN DEFAULT false,
    account BYTEA
);

-- 2. Identity Keys
CREATE TABLE IF NOT EXISTS whatsmeow_identity_keys (
    our_jid TEXT,
    their_id TEXT,
    identity BYTEA,
    PRIMARY KEY (our_jid, their_id),
    FOREIGN KEY (our_jid) REFERENCES whatsmeow_device(jid) ON DELETE CASCADE
);

-- 3. Pre Keys
CREATE TABLE IF NOT EXISTS whatsmeow_pre_keys (
    jid TEXT,
    key_id INTEGER,
    key BYTEA,
    uploaded BOOLEAN,
    PRIMARY KEY (jid, key_id),
    FOREIGN KEY (jid) REFERENCES whatsmeow_device(jid) ON DELETE CASCADE
);

-- 4. Sessions
CREATE TABLE IF NOT EXISTS whatsmeow_sessions (
    our_jid TEXT,
    their_id TEXT,
    session BYTEA,
    PRIMARY KEY (our_jid, their_id),
    FOREIGN KEY (our_jid) REFERENCES whatsmeow_device(jid) ON DELETE CASCADE
);

-- 5. Sender Keys
CREATE TABLE IF NOT EXISTS whatsmeow_sender_keys (
    our_jid TEXT,
    chat_id TEXT,
    sender_id TEXT,
    sender_key BYTEA,
    PRIMARY KEY (our_jid, chat_id, sender_id),
    FOREIGN KEY (our_jid) REFERENCES whatsmeow_device(jid) ON DELETE CASCADE
);

-- 6. App State Sync Keys
CREATE TABLE IF NOT EXISTS whatsmeow_app_state_sync_keys (
    jid TEXT,
    key_id BYTEA,
    key_data BYTEA,
    timestamp BIGINT,
    fingerprint BYTEA,
    PRIMARY KEY (jid, key_id),
    FOREIGN KEY (jid) REFERENCES whatsmeow_device(jid) ON DELETE CASCADE
);

-- 7. App State Version
CREATE TABLE IF NOT EXISTS whatsmeow_app_state_version (
    jid TEXT,
    name TEXT,
    version BIGINT,
    hash BYTEA,
    PRIMARY KEY (jid, name),
    FOREIGN KEY (jid) REFERENCES whatsmeow_device(jid) ON DELETE CASCADE
);

-- 8. App State Mutation MACs
CREATE TABLE IF NOT EXISTS whatsmeow_app_state_mutation_macs (
    jid TEXT,
    name TEXT,
    version BIGINT,
    index_mac BYTEA,
    value_mac BYTEA,
    PRIMARY KEY (jid, name, version, index_mac),
    FOREIGN KEY (jid) REFERENCES whatsmeow_device(jid) ON DELETE CASCADE
);

-- 9. Contacts
CREATE TABLE IF NOT EXISTS whatsmeow_contacts (
    our_jid TEXT,
    their_jid TEXT,
    first_name TEXT,
    full_name TEXT,
    push_name TEXT,
    business_name TEXT,
    PRIMARY KEY (our_jid, their_jid),
    FOREIGN KEY (our_jid) REFERENCES whatsmeow_device(jid) ON DELETE CASCADE
);

-- 10. Chat Settings
CREATE TABLE IF NOT EXISTS whatsmeow_chat_settings (
    our_jid TEXT,
    chat_jid TEXT,
    muted_until BIGINT,
    pinned BOOLEAN,
    archived BOOLEAN,
    PRIMARY KEY (our_jid, chat_jid),
    FOREIGN KEY (our_jid) REFERENCES whatsmeow_device(jid) ON DELETE CASCADE
);

-- 11. Message Secrets
CREATE TABLE IF NOT EXISTS whatsmeow_message_secrets (
    our_jid TEXT,
    chat_jid TEXT,
    sender_jid TEXT,
    message_id TEXT,
    secret BYTEA,
    PRIMARY KEY (our_jid, chat_jid, sender_jid, message_id),
    FOREIGN KEY (our_jid) REFERENCES whatsmeow_device(jid) ON DELETE CASCADE
);

-- 12. Privacy Tokens
CREATE TABLE IF NOT EXISTS whatsmeow_privacy_tokens (
    our_jid TEXT,
    their_jid TEXT,
    token BYTEA,
    timestamp BIGINT,
    PRIMARY KEY (our_jid, their_jid),
    FOREIGN KEY (our_jid) REFERENCES whatsmeow_device(jid) ON DELETE CASCADE
);

-- ================================================
-- Indexes for Performance
-- ================================================
CREATE INDEX IF NOT EXISTS idx_whatsmeow_device_lid ON whatsmeow_device(lid);
CREATE INDEX IF NOT EXISTS idx_whatsmeow_device_initialized ON whatsmeow_device(initialized);
CREATE INDEX IF NOT EXISTS idx_whatsmeow_sessions_their_id ON whatsmeow_sessions(their_id);
CREATE INDEX IF NOT EXISTS idx_whatsmeow_contacts_their_jid ON whatsmeow_contacts(their_jid);

-- ================================================
-- Verification Query
-- ================================================
-- Run this to verify all tables exist:
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = 'public' 
AND table_name LIKE 'whatsmeow_%'
ORDER BY table_name;
