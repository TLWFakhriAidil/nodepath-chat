-- Complete diagnostic check

-- 1. Check server time
SELECT 
    NOW() as server_time,
    NOW()::date as server_date,
    NOW()::time as server_time_only;

-- 2. Check your campaign status
SELECT 
    id,
    title,
    campaign_date,
    scheduled_time,
    status,
    niche,
    target_status,
    message,
    created_at,
    updated_at
FROM campaigns
WHERE title = 'tsst send';

-- 3. Check if any broadcast messages exist
SELECT 
    bm.id,
    bm.status,
    bm.recipient_phone,
    bm.created_at,
    bm.updated_at,
    c.title as campaign_title
FROM broadcast_messages bm
LEFT JOIN campaigns c ON bm.campaign_id = c.id
WHERE bm.created_at > NOW() - INTERVAL '30 minutes'
ORDER BY bm.created_at DESC;

-- 4. Check worker activity
SELECT 
    COUNT(*) as pending_messages,
    MIN(created_at) as oldest_pending,
    MAX(created_at) as newest_pending
FROM broadcast_messages
WHERE status = 'pending';

-- 5. Check your device status
SELECT 
    id,
    device_name,
    status,
    last_seen,
    phone
FROM user_devices
WHERE device_name = 'aqil';