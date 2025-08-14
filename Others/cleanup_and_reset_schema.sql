-- COMPLETE SCHEMA CLEANUP AND RESET
-- This will drop all existing tables and create fresh ones with the latest schema

-- Drop all existing tables (in correct order to handle foreign keys)
DROP TABLE IF EXISTS trigger_process_log CASCADE;
DROP TABLE IF EXISTS device_load_balance CASCADE;
DROP TABLE IF EXISTS sequence_logs CASCADE;
DROP TABLE IF EXISTS broadcast_messages CASCADE;
DROP TABLE IF EXISTS sequence_contacts CASCADE;
DROP TABLE IF EXISTS sequence_steps CASCADE;
DROP TABLE IF EXISTS sequences CASCADE;
DROP TABLE IF EXISTS campaigns CASCADE;
DROP TABLE IF EXISTS leads CASCADE;
DROP TABLE IF EXISTS whatsapp_messages CASCADE;
DROP TABLE IF EXISTS whatsapp_chats CASCADE;
DROP TABLE IF EXISTS message_analytics CASCADE;
DROP TABLE IF EXISTS user_sessions CASCADE;
DROP TABLE IF EXISTS user_devices CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS broadcast_locks CASCADE;

-- Drop whatsmeow tables
DROP TABLE IF EXISTS whatsmeow_device CASCADE;
DROP TABLE IF EXISTS whatsmeow_identity_keys CASCADE;
DROP TABLE IF EXISTS whatsmeow_pre_keys CASCADE;
DROP TABLE IF EXISTS whatsmeow_sender_keys CASCADE;
DROP TABLE IF EXISTS whatsmeow_sessions CASCADE;
DROP TABLE IF EXISTS whatsmeow_app_state_sync_keys CASCADE;
DROP TABLE IF EXISTS whatsmeow_app_state_sync_version CASCADE;
DROP TABLE IF EXISTS whatsmeow_app_state_mutation_macs CASCADE;
DROP TABLE IF EXISTS whatsmeow_contacts CASCADE;
DROP TABLE IF EXISTS whatsmeow_chat_settings CASCADE;
DROP TABLE IF EXISTS whatsmeow_disappearing_timers CASCADE;
DROP TABLE IF EXISTS whatsmeow_history_sync_conversations CASCADE;
DROP TABLE IF EXISTS whatsmeow_history_sync_messages CASCADE;
DROP TABLE IF EXISTS whatsmeow_media_keys CASCADE;
DROP TABLE IF EXISTS whatsmeow_message_secrets CASCADE;
DROP TABLE IF EXISTS whatsmeow_privacy_tokens CASCADE;
DROP TABLE IF EXISTS whatsmeow_newsletter_messages CASCADE;

-- Now we'll create the complete latest schema in the next file