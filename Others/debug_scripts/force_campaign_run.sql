-- Check if timezone fix worked and force campaign to run

-- 1. Check current server time with timezone
SELECT 
    NOW() as current_server_time,
    NOW()::date as server_date,
    NOW()::time as server_time,
    current_setting('TIMEZONE') as database_timezone;

-- 2. Check campaign status and dates
SELECT 
    id,
    title,
    campaign_date,
    scheduled_time,
    status,
    user_id,
    niche,
    target_status,
    updated_at
FROM campaigns
WHERE title = 'tsst send';

-- 3. Force update campaign to today's date and past time
UPDATE campaigns 
SET campaign_date = CURRENT_DATE,  -- Use server's current date
    scheduled_time = '00:00:00',    -- Set to midnight (past time)
    status = 'pending',
    updated_at = NOW()
WHERE title = 'tsst send'
AND status != 'sent';

-- 4. Verify leads exist for this user
SELECT 
    l.id,
    l.phone,
    l.niche,
    l.target_status,
    ud.device_name,
    ud.user_id,
    ud.status as device_status
FROM leads l
JOIN user_devices ud ON l.device_id = ud.id
WHERE l.niche LIKE '%VITAC%'
AND l.target_status = 'customer';

-- 5. Manually create a broadcast message to test worker
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
JOIN leads l ON l.niche LIKE '%' || c.niche || '%'
JOIN user_devices ud ON l.device_id = ud.id AND ud.user_id = c.user_id
WHERE c.title = 'tsst send'
AND l.target_status = c.target_status
AND NOT EXISTS (
    SELECT 1 FROM broadcast_messages bm 
    WHERE bm.campaign_id = c.id 
    AND bm.recipient_phone = l.phone
)
LIMIT 1;

-- 6. Check if broadcast message was created
SELECT 
    bm.id,
    bm.status,
    bm.recipient_phone,
    bm.created_at,
    c.title as campaign_title,
    ud.device_name
FROM broadcast_messages bm
JOIN campaigns c ON bm.campaign_id = c.id
JOIN user_devices ud ON bm.device_id = ud.id
WHERE c.title = 'tsst send'
ORDER BY bm.created_at DESC;