-- Check for devices that might appear as duplicates
-- This will show all devices grouped by user and similar names

-- 1. Show all devices with their key fields
SELECT 
    id,
    user_id,
    device_name,
    jid,
    platform,
    status,
    created_at
FROM user_devices
ORDER BY user_id, device_name, created_at;

-- 2. Find devices with similar names for same user
SELECT 
    user_id,
    device_name,
    COUNT(*) as device_count,
    STRING_AGG(id::text, ', ') as device_ids,
    STRING_AGG(jid, ', ') as jids
FROM user_devices
GROUP BY user_id, device_name
HAVING COUNT(*) > 1
ORDER BY device_count DESC;

-- 3. Find devices created via webhook (likely have timestamp in name)
SELECT 
    id,
    user_id,
    device_name,
    jid,
    platform,
    created_at
FROM user_devices
WHERE device_name LIKE 'Device-%'
ORDER BY created_at DESC
LIMIT 20;

-- 4. Check for non-UUID JIDs (platform devices)
SELECT 
    id,
    device_name,
    jid,
    platform,
    CASE 
        WHEN jid ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$' 
        THEN 'UUID'
        ELSE 'Non-UUID (Platform/External)'
    END as jid_type
FROM user_devices
ORDER BY created_at DESC;