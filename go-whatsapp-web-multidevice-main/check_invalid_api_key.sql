-- Check for devices with invalid API key errors in broadcast_messages
SELECT DISTINCT
    bm.device_id,
    ud.device_name,
    ud.platform,
    ud.status as device_status,
    COUNT(DISTINCT bm.id) as error_count,
    MIN(bm.created_at) as first_error,
    MAX(bm.created_at) as last_error,
    bm.error_message
FROM broadcast_messages bm
LEFT JOIN user_devices ud ON ud.id = bm.device_id
WHERE bm.error_message LIKE '%invalid api key%' 
   OR bm.error_message LIKE '%Invalid API%'
   OR bm.error_message LIKE '%API key%'
GROUP BY bm.device_id, ud.device_name, ud.platform, ud.status, bm.error_message
ORDER BY error_count DESC;

-- Get sample of recent messages with API key errors
SELECT 
    bm.id,
    bm.device_id,
    ud.device_name,
    ud.platform,
    bm.recipient_phone,
    bm.status,
    bm.error_message,
    bm.created_at
FROM broadcast_messages bm
LEFT JOIN user_devices ud ON ud.id = bm.device_id
WHERE bm.error_message LIKE '%invalid api key%' 
   OR bm.error_message LIKE '%Invalid API%'
   OR bm.error_message LIKE '%API key%'
ORDER BY bm.created_at DESC
LIMIT 20;

-- Count total messages affected
SELECT 
    COUNT(*) as total_messages_with_api_error,
    COUNT(DISTINCT device_id) as total_devices_affected
FROM broadcast_messages
WHERE error_message LIKE '%invalid api key%' 
   OR error_message LIKE '%Invalid API%'
   OR error_message LIKE '%API key%';
