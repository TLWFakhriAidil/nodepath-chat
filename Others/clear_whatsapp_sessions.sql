-- Clear all WhatsApp session data
-- WARNING: This will log out all WhatsApp connections

-- Clear whatsmeow tables (these store WhatsApp session data)
DROP TABLE IF EXISTS whatsmeow_app_state_mutation_macs CASCADE;
DROP TABLE IF EXISTS whatsmeow_app_state_sync_keys CASCADE;
DROP TABLE IF EXISTS whatsmeow_app_state_version CASCADE;
DROP TABLE IF EXISTS whatsmeow_chat_settings CASCADE;
DROP TABLE IF EXISTS whatsmeow_contacts CASCADE;
DROP TABLE IF EXISTS whatsmeow_device CASCADE;
DROP TABLE IF EXISTS whatsmeow_identity_keys CASCADE;
DROP TABLE IF EXISTS whatsmeow_message_secrets CASCADE;
DROP TABLE IF EXISTS whatsmeow_pre_keys CASCADE;
DROP TABLE IF EXISTS whatsmeow_privacy_tokens CASCADE;
DROP TABLE IF EXISTS whatsmeow_sender_keys CASCADE;
DROP TABLE IF EXISTS whatsmeow_sessions CASCADE;

-- Update all devices to offline status
UPDATE user_devices 
SET status = 'offline', 
    jid = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE status = 'online';

-- Show current devices
SELECT id, device_name, phone, jid, status 
FROM user_devices 
WHERE user_id = 'de078f16-3266-4ab3-8153-a248b015228f';
