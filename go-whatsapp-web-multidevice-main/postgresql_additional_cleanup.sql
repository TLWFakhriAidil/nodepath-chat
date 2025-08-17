-- PostgreSQL Additional Cleanup Commands
-- Use these if you need to free up more space (currently at 121 MB)
-- WARNING: These will delete WhatsApp session data!

-- ============================================
-- CURRENT STORAGE BREAKDOWN:
-- WhatsApp Session Data: 73.9 MB (61.1%)
-- WhatsApp Chat/Messages: 39 MB (32.2%)
-- Application Data: 0.9 MB (0.7%)
-- ============================================

-- 1. CHECK CURRENT SIZES BEFORE CLEANUP
SELECT 
    tablename,
    pg_size_pretty(pg_total_relation_size('public.'||tablename)) as size,
    pg_total_relation_size('public.'||tablename)/1024/1024 as size_mb
FROM pg_tables
WHERE schemaname = 'public' 
AND tablename IN ('whatsmeow_message_secrets', 'whatsapp_messages', 'whatsmeow_contacts', 'whatsapp_chats')
ORDER BY size_mb DESC;

-- 2. CLEAN OLD MESSAGE ENCRYPTION KEYS (Potential: -20 MB)
-- This removes encryption keys for inactive devices
DELETE FROM whatsmeow_message_secrets 
WHERE jid IN (
    SELECT jid FROM user_devices 
    WHERE status != 'online' 
    OR last_seen < NOW() - INTERVAL '30 days'
);

-- 3. ARCHIVE OLD MESSAGES (Potential: -15 MB)
-- Delete messages older than 30 days
DELETE FROM whatsapp_messages 
WHERE created_at < NOW() - INTERVAL '30 days';

-- Alternative: Keep only last 7 days
-- DELETE FROM whatsapp_messages 
-- WHERE created_at < NOW() - INTERVAL '7 days';

-- 4. CLEAN DUPLICATE CONTACTS (Potential: -5 MB)
-- Remove duplicate contacts, keeping the most recent
WITH duplicates AS (
    SELECT id, 
           ROW_NUMBER() OVER (PARTITION BY jid ORDER BY id DESC) as rn
    FROM whatsmeow_contacts
)
DELETE FROM whatsmeow_contacts
WHERE id IN (SELECT id FROM duplicates WHERE rn > 1);

-- 5. CLEAN OLD CHATS (Potential: -5 MB)
-- Remove chats with no recent activity
DELETE FROM whatsapp_chats
WHERE last_message_time < NOW() - INTERVAL '60 days'
AND chat_jid NOT IN (
    SELECT DISTINCT chat_jid 
    FROM whatsapp_messages 
    WHERE created_at > NOW() - INTERVAL '60 days'
);

-- 6. CLEAN ORPHANED SESSION DATA
-- Remove session data for deleted devices
DELETE FROM whatsmeow_sessions
WHERE jid NOT IN (SELECT jid FROM user_devices);

DELETE FROM whatsmeow_identity_keys
WHERE jid NOT IN (SELECT jid FROM user_devices);

DELETE FROM whatsmeow_sender_keys
WHERE chat_jid NOT IN (SELECT DISTINCT chat_jid FROM whatsapp_chats);

-- 7. CLEAN APP STATE DATA FOR INACTIVE DEVICES
DELETE FROM whatsmeow_app_state_mutation_macs
WHERE jid NOT IN (SELECT jid FROM user_devices WHERE status = 'online');

-- 8. VACUUM TO RECLAIM SPACE
VACUUM FULL;

-- 9. CHECK SIZES AFTER CLEANUP
SELECT 
    'Total DB Size' as metric,
    pg_size_pretty(pg_database_size(current_database())) as size
UNION ALL
SELECT 
    'After Cleanup - ' || tablename,
    pg_size_pretty(pg_total_relation_size('public.'||tablename))
FROM pg_tables
WHERE schemaname = 'public' 
AND pg_total_relation_size('public.'||tablename) > 1024*1024
ORDER BY 1;

-- ============================================
-- EXTREME CLEANUP (Use with caution!)
-- This will reset most WhatsApp data
-- ============================================

-- Nuclear option: Clear all WhatsApp message history
-- TRUNCATE TABLE whatsapp_messages CASCADE;
-- TRUNCATE TABLE whatsapp_chats CASCADE;

-- Clear all but active session encryption
-- TRUNCATE TABLE whatsmeow_message_secrets CASCADE;

-- Clear all contacts (will resync on next connection)
-- TRUNCATE TABLE whatsmeow_contacts CASCADE;

-- ============================================
-- EXPECTED RESULTS:
-- Conservative cleanup: 121 MB -> 80-90 MB
-- Aggressive cleanup: 121 MB -> 60-70 MB
-- Extreme cleanup: 121 MB -> 20-30 MB
-- ============================================
