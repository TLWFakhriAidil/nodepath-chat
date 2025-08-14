-- Check broadcast messages status
SELECT status, COUNT(*) as count 
FROM broadcast_messages 
GROUP BY status;

-- Check pending messages
SELECT id, device_id, recipient_phone, recipient_name, status, 
       created_at, scheduled_at, error_message
FROM broadcast_messages 
WHERE status IN ('pending', 'queued')
ORDER BY created_at DESC
LIMIT 10;

-- Check if devices are online
SELECT ud.id, ud.device_name, ud.status, ud.phone
FROM user_devices ud
WHERE EXISTS (
    SELECT 1 FROM broadcast_messages bm 
    WHERE bm.device_id = ud.id 
    AND bm.status IN ('pending', 'queued')
)
ORDER BY ud.device_name;