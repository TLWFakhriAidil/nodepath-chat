-- Debug SQL queries to understand the sequence leads issue
-- Run these queries to diagnose why we're seeing 0 unique leads

-- 1. Check if there are any broadcast messages for this sequence
SELECT COUNT(*) as total_messages
FROM broadcast_messages
WHERE sequence_id = 'YOUR_SEQUENCE_ID';

-- 2. Check unique phone+device combinations
SELECT COUNT(DISTINCT CONCAT(recipient_phone, '|', device_id)) as unique_leads
FROM broadcast_messages
WHERE sequence_id = 'YOUR_SEQUENCE_ID';

-- 3. Check if recipient_phone or device_id might be NULL
SELECT 
    COUNT(*) as total_messages,
    SUM(CASE WHEN recipient_phone IS NULL THEN 1 ELSE 0 END) as null_phones,
    SUM(CASE WHEN device_id IS NULL THEN 1 ELSE 0 END) as null_devices,
    SUM(CASE WHEN recipient_phone = '' THEN 1 ELSE 0 END) as empty_phones,
    SUM(CASE WHEN device_id = '' THEN 1 ELSE 0 END) as empty_devices
FROM broadcast_messages
WHERE sequence_id = 'YOUR_SEQUENCE_ID';

-- 4. Sample some actual data
SELECT 
    recipient_phone,
    device_id,
    sequence_stepid,
    status,
    COUNT(*) as count
FROM broadcast_messages
WHERE sequence_id = 'YOUR_SEQUENCE_ID'
GROUP BY recipient_phone, device_id, sequence_stepid, status
LIMIT 10;

-- 5. Check step-wise breakdown
SELECT 
    sequence_stepid,
    COUNT(*) as total_messages,
    COUNT(DISTINCT recipient_phone) as unique_phones,
    COUNT(DISTINCT device_id) as unique_devices,
    COUNT(DISTINCT CONCAT(recipient_phone, '|', device_id)) as unique_combinations
FROM broadcast_messages
WHERE sequence_id = 'YOUR_SEQUENCE_ID'
GROUP BY sequence_stepid;

-- 6. Check if the issue is with date filtering
SELECT 
    DATE(scheduled_at) as scheduled_date,
    COUNT(*) as messages_count,
    COUNT(DISTINCT CONCAT(recipient_phone, '|', device_id)) as unique_leads
FROM broadcast_messages
WHERE sequence_id = 'YOUR_SEQUENCE_ID'
GROUP BY DATE(scheduled_at)
ORDER BY scheduled_date DESC;
