-- ================================================
-- Clean WhatsApp Session Tables
-- ================================================
-- This script removes all whatsmeow tables that were added
-- for session storage but are causing 502 errors

-- Drop all whatsmeow tables (CASCADE will handle foreign keys)
DROP TABLE IF EXISTS whatsmeow_message_secrets CASCADE;
DROP TABLE IF EXISTS whatsmeow_contacts CASCADE;
DROP TABLE IF EXISTS whatsmeow_chat_settings CASCADE;
DROP TABLE IF EXISTS whatsmeow_app_state_mutation_macs CASCADE;
DROP TABLE IF EXISTS whatsmeow_app_state_version CASCADE;
DROP TABLE IF EXISTS whatsmeow_app_state_sync_keys CASCADE;
DROP TABLE IF EXISTS whatsmeow_sender_keys CASCADE;
DROP TABLE IF EXISTS whatsmeow_sessions CASCADE;
DROP TABLE IF EXISTS whatsmeow_pre_keys CASCADE;
DROP TABLE IF EXISTS whatsmeow_identity_keys CASCADE;
DROP TABLE IF EXISTS whatsmeow_device CASCADE;

-- List remaining tables to verify cleanup
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = 'public' 
AND table_name LIKE 'whatsmeow%'
ORDER BY table_name;

-- Verify core tables are still intact
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = 'public' 
AND table_name IN (
    'users',
    'user_devices', 
    'user_sessions',
    'campaigns',
    'sequences',
    'leads',
    'broadcast_messages',
    'whatsapp_chats',
    'whatsapp_messages'
)
ORDER BY table_name;