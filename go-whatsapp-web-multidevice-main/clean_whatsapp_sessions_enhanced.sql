-- Enhanced WhatsApp Session Cleanup Script
-- This script safely cleans WhatsApp session data while handling foreign key constraints

-- Function to clean WhatsApp session for a specific JID
CREATE OR REPLACE FUNCTION clean_whatsapp_session_enhanced(p_jid TEXT)
RETURNS void AS $$
BEGIN
    -- Temporarily disable foreign key checks
    SET session_replication_role = 'replica';
    
    -- Clean all tables that might contain the JID
    -- Using DELETE with error handling for each table
    
    -- Privacy tokens
    DELETE FROM whatsmeow_privacy_tokens WHERE our_jid = p_jid;
    
    -- Portal related tables (in order)
    DELETE FROM whatsmeow_portal_message_part 
    WHERE message_id IN (
        SELECT id FROM whatsmeow_portal_message 
        WHERE portal_jid IN (SELECT jid FROM whatsmeow_portal WHERE receiver = p_jid)
    );
    
    DELETE FROM whatsmeow_portal_reaction 
    WHERE portal_jid IN (SELECT jid FROM whatsmeow_portal WHERE receiver = p_jid);
    
    DELETE FROM whatsmeow_portal_message 
    WHERE portal_jid IN (SELECT jid FROM whatsmeow_portal WHERE receiver = p_jid);
    
    DELETE FROM whatsmeow_portal_backfill_queue 
    WHERE portal_jid IN (SELECT jid FROM whatsmeow_portal WHERE receiver = p_jid);
    
    DELETE FROM whatsmeow_portal_backfill 
    WHERE portal_jid IN (SELECT jid FROM whatsmeow_portal WHERE receiver = p_jid);
    
    DELETE FROM whatsmeow_media_backfill_requests 
    WHERE user_jid = p_jid OR portal_jid = p_jid;
    
    DELETE FROM whatsmeow_portal WHERE receiver = p_jid OR jid = p_jid;
    
    -- Message and chat related
    DELETE FROM whatsmeow_message_secrets WHERE chat_jid = p_jid OR sender_id = p_jid;
    DELETE FROM whatsmeow_history_syncs WHERE device_jid = p_jid;
    
    -- Group related
    DELETE FROM whatsmeow_group_participants WHERE group_jid = p_jid OR jid = p_jid;
    DELETE FROM whatsmeow_groups WHERE jid = p_jid;
    
    -- Contact and settings
    DELETE FROM whatsmeow_disappearing_timers WHERE jid = p_jid;
    DELETE FROM whatsmeow_chat_settings WHERE jid = p_jid;
    DELETE FROM whatsmeow_contacts WHERE jid = p_jid OR our_jid = p_jid;
    
    -- App state
    DELETE FROM whatsmeow_app_state_mutation_macs WHERE jid = p_jid;
    DELETE FROM whatsmeow_app_state_sync_keys WHERE jid = p_jid;
    DELETE FROM whatsmeow_app_state_version WHERE jid = p_jid;
    
    -- Keys and sessions
    DELETE FROM whatsmeow_sender_keys WHERE chat_id = p_jid OR sender_id = p_jid;
    DELETE FROM whatsmeow_sessions WHERE their_id = p_jid;
    DELETE FROM whatsmeow_pre_keys WHERE jid = p_jid;
    DELETE FROM whatsmeow_identity_keys WHERE their_id = p_jid;
    
    -- Finally, delete the device
    DELETE FROM whatsmeow_device WHERE jid = p_jid;
    
    -- Re-enable foreign key checks
    SET session_replication_role = 'origin';
    
EXCEPTION
    WHEN OTHERS THEN
        -- Re-enable foreign key checks even on error
        SET session_replication_role = 'origin';
        RAISE NOTICE 'Error cleaning session for %: %', p_jid, SQLERRM;
END;
$$ LANGUAGE plpgsql;

-- Clean all orphaned WhatsApp sessions
CREATE OR REPLACE FUNCTION clean_orphaned_whatsapp_sessions()
RETURNS void AS $$
DECLARE
    v_jid TEXT;
BEGIN
    -- Find all JIDs in whatsmeow tables that don't have a corresponding device
    FOR v_jid IN 
        SELECT DISTINCT jid FROM (
            SELECT jid FROM whatsmeow_device
            UNION
            SELECT our_jid as jid FROM whatsmeow_contacts
            UNION
            SELECT jid FROM whatsmeow_app_state_version
        ) all_jids
        WHERE jid NOT IN (SELECT jid FROM user_devices WHERE jid IS NOT NULL)
    LOOP
        PERFORM clean_whatsapp_session_enhanced(v_jid);
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Usage examples:
-- Clean a specific device session:
-- SELECT clean_whatsapp_session_enhanced('60146674397:54@s.whatsapp.net');

-- Clean all orphaned sessions:
-- SELECT clean_orphaned_whatsapp_sessions();
