-- Manual campaign execution test
-- This will show exactly what's happening

-- 1. Show campaign details
SELECT 
    id,
    user_id,
    title,
    niche,
    target_status,
    message,
    status
FROM campaigns
WHERE title = 'tsst send';

-- 2. Show lead details with user comparison
SELECT 
    l.id,
    l.name,
    l.phone,
    l.niche,
    l.target_status,
    l.user_id as lead_user_id,
    l.device_id,
    ud.user_id as device_user_id,
    c.user_id as campaign_user_id,
    CASE 
        WHEN l.user_id = c.user_id THEN 'MATCH'
        WHEN ud.user_id = c.user_id THEN 'MATCH_BY_DEVICE'
        ELSE 'NO_MATCH'
    END as user_match
FROM leads l
JOIN user_devices ud ON l.device_id = ud.id
CROSS JOIN campaigns c
WHERE l.phone = '60108924904'
AND c.title = 'tsst send';

-- 3. Get all leads for your user's devices
SELECT 
    l.*,
    ud.device_name
FROM leads l
JOIN user_devices ud ON l.device_id = ud.id
WHERE ud.device_name = 'aqil'
AND l.niche LIKE '%VITAC%'
AND l.target_status = 'customer';

-- 4. Force create broadcast message
INSERT INTO broadcast_messages (
    user_id,
    device_id,
    campaign_id,
    recipient_phone,
    type,
    content,
    status,
    scheduled_at,
    created_at,
    updated_at
)
SELECT 
    c.user_id,
    l.device_id,
    c.id,
    l.phone,
    'text',
    c.message,
    'pending',
    NOW(),
    NOW(),
    NOW()
FROM campaigns c
CROSS JOIN leads l
WHERE c.title = 'tsst send'
AND l.phone = '60108924904'
AND NOT EXISTS (
    SELECT 1 FROM broadcast_messages bm 
    WHERE bm.campaign_id = c.id 
    AND bm.recipient_phone = l.phone
);

-- 5. Check result
SELECT COUNT(*) as messages_created
FROM broadcast_messages
WHERE recipient_phone = '60108924904';