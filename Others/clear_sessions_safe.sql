-- Handle WhatsApp session tables
-- These tables are auto-created by whatsmeow library
-- We'll truncate them if they exist to prevent issues

DO $$
BEGIN
    -- Check if whatsmeow_device table exists
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'whatsmeow_device') THEN
        -- Clear all session data
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
        RAISE NOTICE 'WhatsApp session tables cleared';
    ELSE
        RAISE NOTICE 'WhatsApp session tables do not exist yet';
    END IF;
END $$;
