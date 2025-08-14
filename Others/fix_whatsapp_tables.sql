-- Recreate WhatsApp tables that were dropped by Clear All Sessions
-- This script creates all necessary whatsmeow tables

-- Device table
CREATE TABLE IF NOT EXISTS whatsmeow_device (
    jid text PRIMARY KEY,
    registration_id bigint NOT NULL,
    noise_key bytea NOT NULL,
    identity_key bytea NOT NULL,
    signed_pre_key bytea NOT NULL,
    signed_pre_key_id integer NOT NULL,
    signed_pre_key_sig bytea NOT NULL,
    adv_key bytea,
    adv_details bytea,
    adv_account_sig bytea,
    adv_device_sig bytea,
    platform text DEFAULT '',
    business_name text DEFAULT '',
    push_name text DEFAULT '',
    initialized boolean DEFAULT false,
    lid text,
    facebook_uuid uuid,
    account text
);

-- Sessions table
CREATE TABLE IF NOT EXISTS whatsmeow_sessions (
    our_jid text,
    their_id text,
    session bytea,
    PRIMARY KEY (our_jid, their_id),
    FOREIGN KEY (our_jid) REFERENCES whatsmeow_device(jid) ON DELETE CASCADE
);

-- Identity keys table
CREATE TABLE IF NOT EXISTS whatsmeow_identity_keys (
    our_jid text,
    their_id text,
    identity bytea,
    PRIMARY KEY (our_jid, their_id),
    FOREIGN KEY (our_jid) REFERENCES whatsmeow_device(jid) ON DELETE CASCADE
);

-- Pre-keys table
CREATE TABLE IF NOT EXISTS whatsmeow_pre_keys (
    jid text,
    key_id integer,
    key bytea,
    uploaded boolean DEFAULT false,
    PRIMARY KEY (jid, key_id),
    FOREIGN KEY (jid) REFERENCES whatsmeow_device(jid) ON DELETE CASCADE
);

-- Sender keys table
CREATE TABLE IF NOT EXISTS whatsmeow_sender_keys (
    our_jid text,
    chat_id text,
    sender_id text,
    sender_key bytea NOT NULL,
    PRIMARY KEY (our_jid, chat_id, sender_id),
    FOREIGN KEY (our_jid) REFERENCES whatsmeow_device(jid) ON DELETE CASCADE
);

-- App state sync keys table
CREATE TABLE IF NOT EXISTS whatsmeow_app_state_sync_keys (
    jid text,
    key_id byt