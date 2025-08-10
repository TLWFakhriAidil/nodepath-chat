-- STEP BY STEP DEBUG

-- 1. Check if broadcast messages exist for your campaign
SELECT 
    bm.*,
    c.title as campaign_title
FROM broadcast_messages bm
JOIN campaigns c ON bm.campaign_id = c.id
WHERE c.title = 'tsst send';

-- 2. If no messages above, check why campaign isn't creating messages
SELECT 
    c.id as campaign_id,
    c.user_id as campaign_user_id,
    c.niche as campaign_niche,
    c.target_status as campaign_target,
    c.status as campaign_status,
    l.phone as lead_phone,
    l.niche as lead_niche,
    l.target_status as lead_target,
    l.user_id as lead_user_id,
    ud.user_id as device_user_id,
    CASE 
        WHEN l.user_id = c.user_id THEN 'USER_MATCH'
        WHEN ud.user_id = c.user_id THEN 'DEVICE_USER_MATCH'
        ELSE 'NO_MATCH'
    END as match_status
FROM campaigns c
LEFT JOIN leads l ON l.niche LIKE '%' || c.niche || '%' 
    AND l.target_status = c.target_status
LEFT JOIN user_devices ud ON l.device_id = ud.id
WHERE c.title = 'tsst send';

-- 3. Force create a test message to trigger worker
INSERT INTO broadcast_messages (
    user_id,
    device_id,
    recipient_phone,
    type,
    content,
    status,
    scheduled_at,
    created_at,
    updated_at
) VALUES (
    'de078f16-3266-4ab3-8153-a248b015228f',  -- Your user ID
    '2de48db2-f1ab-4d81-8a26-58b01df75bdf',  -- aqil device ID
    '60108924904',  -- Your phone
    'text',
    'DEBUG: Testing worker system at ' || NOW()::text,
    'pending',
    NOW(),
    NOW(),
    NOW()
)
ON CONFLICT DO NOTHING;

-- 4. Check pending messages
SELECT 
    id,
    status,
    recipient_phone,
    content,
    created_at,
    device_id
FROM broadcast_messages
WHERE status = 'pending'
ORDER BY created_at DESC
LIMIT 10;

-- 5. Check Redis queue (if messages exist but worker not picking up)
-- This shows if it's a worker issue or message creation issue