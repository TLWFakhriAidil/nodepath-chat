-- Debug what's actually happening with the campaign

-- 1. Check the exact values in your campaign
SELECT 
    id,
    title,
    campaign_date,
    scheduled_time,
    scheduled_time::text as time_text,
    status,
    CASE 
        WHEN scheduled_time IS NULL THEN 'IS_NULL'
        WHEN scheduled_time::text = '' THEN 'IS_EMPTY'
        WHEN scheduled_time::text = '00:00:00' THEN 'IS_MIDNIGHT'
        ELSE 'HAS_TIME: ' || scheduled_time::text
    END as time_check
FROM campaigns
WHERE title = 'tsst send';

-- 2. Force the campaign to a state that will definitely trigger
UPDATE campaigns 
SET status = 'pending',
    campaign_date = '2025-06-27',  -- Today for server
    scheduled_time = NULL           -- NULL should trigger immediately
WHERE title = 'tsst send';

-- 3. Create a simple test campaign that will definitely work
INSERT INTO campaigns (
    user_id, 
    title, 
    niche, 
    target_status, 
    message, 
    campaign_date, 
    scheduled_time,
    min_delay_seconds, 
    max_delay_seconds, 
    status, 
    created_at, 
    updated_at
) VALUES (
    'de078f16-3266-4ab3-8153-a248b015228f',
    'simple_test_' || date_part('minute', NOW()),
    'VITAC',
    'customer', 
    'TEST MESSAGE - this should work!',
    '2025-06-27',    -- Server date
    NULL,            -- No time = immediate
    5,
    10,
    'pending',
    NOW(),
    NOW()
);

-- 4. Check if there are ANY pending messages in broadcast queue
SELECT COUNT(*) as pending_count
FROM broadcast_messages
WHERE status = 'pending';

-- 5. Manually queue a message to test the worker system
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
    'de078f16-3266-4ab3-8153-a248b015228f',
    '2de48db2-f1ab-4d81-8a26-58b01df75bdf',
    '60108924904',
    'text',
    'DIRECT TEST: Time ' || NOW()::time,
    'pending',
    NOW(),
    NOW(),
    NOW()
);

-- This will tell us if:
-- a) The campaign trigger is working but not creating messages
-- b) Messages are being created but workers aren't picking them up
-- c) The whole system is not working