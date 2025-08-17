-- ================================================
-- Clear Data from WhatsApp Session Tables
-- ================================================
-- This script empties all whatsmeow tables without dropping them
-- TRUNCATE is faster than DELETE and resets any auto-increment counters

-- Clear all data from whatsmeow tables
TRUNCATE TABLE whatsmeow_message_secrets CASCADE;
TRUNCATE TABLE whatsmeow_contacts CASCADE;
TRUNCATE TABLE whatsmeow_chat_settings CASCADE;
TRUNCATE TABLE whatsmeow_app_state_mutation_macs CASCADE;
TRUNCATE TABLE whatsmeow_app_state_version CASCADE;
TRUNCATE TABLE whatsmeow_app_state_sync_keys CASCADE;
TRUNCATE TABLE whatsmeow_sender_keys CASCADE;
TRUNCATE TABLE whatsmeow_sessions CASCADE;
TRUNCATE TABLE whatsmeow_pre_keys CASCADE;
TRUNCATE TABLE whatsmeow_identity_keys CASCADE;
TRUNCATE TABLE whatsmeow_device CASCADE;

-- Verify tables are empty
SELECT 'whatsmeow_device' as table_name, COUNT(*) as row_count FROM whatsmeow_device
UNION ALL
SELECT 'whatsmeow_identity_keys', COUNT(*) FROM whatsmeow_identity_keys
UNION ALL
SELECT 'whatsmeow_pre_keys', COUNT(*) FROM whatsmeow_pre_keys
UNION ALL
SELECT 'whatsmeow_sessions', COUNT(*) FROM whatsmeow_sessions
UNION ALL
SELECT 'whatsmeow_sender_keys', COUNT(*) FROM whatsmeow_sender_keys
UNION ALL
SELECT 'whatsmeow_app_state_sync_keys', COUNT(*) FROM whatsmeow_app_state_sync_keys
UNION ALL
SELECT 'whatsmeow_app_state_version', COUNT(*) FROM whatsmeow_app_state_version
UNION ALL
SELECT 'whatsmeow_app_state_mutation_macs', COUNT(*) FROM whatsmeow_app_state_mutation_macs
UNION ALL
SELECT 'whatsmeow_message_secrets', COUNT(*) FROM whatsmeow_message_secrets
UNION ALL
SELECT 'whatsmeow_contacts', COUNT(*) FROM whatsmeow_contacts
UNION ALL
SELECT 'whatsmeow_chat_settings', COUNT(*) FROM whatsmeow_chat_settings;