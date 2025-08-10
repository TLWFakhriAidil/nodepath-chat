-- Simple delete all records from sequence_contacts and broadcast_messages

-- Delete all broadcast_messages
DELETE FROM broadcast_messages;

-- Delete all sequence_contacts  
DELETE FROM sequence_contacts;

-- Show results
SELECT 'broadcast_messages' as table_name, COUNT(*) as record_count FROM broadcast_messages
UNION ALL
SELECT 'sequence_contacts' as table_name, COUNT(*) as record_count FROM sequence_contacts;
