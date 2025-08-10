-- Check broadcast_messages table structure and data

-- 1. Show table structure
DESCRIBE broadcast_messages;

-- 2. Check all unique statuses in the table
SELECT DISTINCT status, COUNT(*) as count
FROM broadcast_messages
GROUP BY status
ORDER BY count DESC;

-- 3. Check last 20 messages with all important columns
SELECT 
    id,
    device_id,
    campaign_id,
    sequence_id,
    recipient_phone,
    recipient_name,
    message_type,
    status,
    error_message,
    scheduled_at,
    sent_at,
    created_at,
    updated_at
FROM broadcast_messages
ORDER BY created_at DESC
LIMIT 20;

-- 4. Check messages that should be processed (pending with past scheduled time)
SELECT 
    id,
    device_id,
    recipient_phone,
    status,
    scheduled_at,
    created_at,
    CASE 
        WHEN scheduled_at IS NULL THEN 'No schedule'
        WHEN scheduled_at <= NOW() THEN 'Should process'
        ELSE 'Future'
    END as schedule_status
FROM broadcast_messages
WHERE status = 'pending'
ORDER BY created_at DESC
LIMIT 10;

-- 5. Count by status and message type
SELECT 
    status,
    message_type,
    COUNT(*) as count
FROM broadcast_messages
GROUP BY status, message_type
ORDER BY status, message_type;

-- 6. Check if recipient_name is populated
SELECT 
    COUNT(*) as total,
    SUM(CASE WHEN recipient_name IS NULL THEN 1 ELSE 0 END) as null_names,
    SUM(CASE WHEN recipient_name = '' THEN 1 ELSE 0 END) as empty_names,
    SUM(CASE WHEN recipient_name IS NOT NULL AND recipient_name != '' THEN 1 ELSE 0 END) as has_names
FROM broadcast_messages
WHERE status = 'pending';