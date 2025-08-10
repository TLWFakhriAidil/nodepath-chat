-- Fix for WhatsApp session cleanup
-- This script properly clears all WhatsApp session data

-- Function to safely clear WhatsApp session for a device
CREATE OR REPLACE FUNCTION clear_whatsapp_session(device_id TEXT)
RETURNS void AS $$
DECLARE
    device_jid TEXT;
BEGIN
    -- First try to find the JID from user_devices
    SELECT jid INTO device_jid FROM user_devices WHERE id = device_id;
    
    -- If we have a JID, use it to clear sessions
    IF device_jid IS NOT NULL THEN
        -- Delete in correct order to avoid foreign key violations
        DELETE FROM whatsmeow_app_state_mutation_macs WHERE jid = device_jid;
        DELETE FROM whatsmeow_app_state_sync_keys WHERE jid = device_jid;
        DELETE FROM whatsmeow_app_state_version WHERE jid = device_jid;
        DELETE FROM whatsmeow_chat_settings WHERE jid = device_jid;
        DELETE FROM whatsmeow_contacts WHERE jid = device_jid;
        DELETE FROM whatsmeow_disappearing_timers WHERE jid = device_jid;
        DELETE FROM whatsmeow_group_participants WHERE group_jid IN (SELECT jid FROM whatsmeow_groups WHERE jid = device_jid);
        DELETE FROM whatsmeow_groups WHERE jid = device_jid;
        DELETE FROM whatsmeow_history_syncs WHERE device_jid = device_jid;
        DELETE FROM whatsmeow_media_backfill_requests WHERE chat_jid = device_jid;
        DELETE FROM whatsmeow_message_secrets WHERE chat_jid = device_jid;
        DELETE FROM whatsmeow_portal_backfill WHERE portal_jid IN (SELECT jid FROM whatsmeow_portal WHERE receiver = device_jid);
        DELETE FROM whatsmeow_portal_backfill_queue WHERE portal_jid IN (SELECT jid FROM whatsmeow_portal WHERE receiver = device_jid);
        DELETE FROM whatsmeow_portal_message WHERE portal_jid IN (SELECT jid FROM whatsmeow_portal WHERE receiver = device_jid);
        DELETE FROM whatsmeow_portal_message_part WHERE message_id IN (SELECT id FROM whatsmeow_portal_message WHERE portal_jid IN (SELECT jid FROM whatsmeow_portal WHERE receiver = device_jid));
        DELETE FROM whatsmeow_portal_reaction WHERE portal_jid IN (SELECT jid FROM whatsmeow_portal WHERE receiver = device_jid);
        DELETE FROM whatsmeow_portal WHERE receiver = device_jid;
        DELETE FROM whatsmeow_privacy_settings WHERE jid = device_jid;
        DELETE FROM whatsmeow_sender_keys WHERE our_jid = device_jid;
        DELETE FROM whatsmeow_sessions WHERE our_jid = device_jid;
        DELETE FROM whatsmeow_pre_keys WHERE jid = device_jid;
        DELETE FROM whatsmeow_identity_keys WHERE our_jid = device_jid;
        DELETE FROM whatsmeow_device WHERE jid = device_jid;
    END IF;
    
    -- Also try to delete by device_id directly (in case JID is stored as device_id)
    DELETE FROM whatsmeow_device WHERE jid = device_id;
    
    -- Update the device status
    UPDATE user_devices SET status = 'offline', phone = NULL, jid = NULL WHERE id = device_id;
    
EXCEPTION
    WHEN OTHERS THEN
        -- Log error but don't fail
        RAISE NOTICE 'Error clearing session: %', SQLERRM;
END;
$$ LANGUAGE plpgsql;
