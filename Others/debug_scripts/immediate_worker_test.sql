-- IMMEDIATE TEST: Create a broadcast message to test workers

-- 1. Create a test broadcast message directly
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
    'de078f16-3266-4ab3-8153-a248b015228f',  -- Your user ID from logs
    '2de48db2-f1ab-4d81-8a26-58b01df75bdf',  -- Your device ID
    '60108924904',                            -- Your test phone
    'text',
    'Worker Test: If you receive this, workers are functioning!',
    'pending',
    NOW(),
    NOW(),
    NOW()
);

-- 2. Check if message was created
SELECT 
    id,
    status,
    recipient_phone,
    content,
    created_at
FROM broadcast_messages 
WHERE user_id = 'de078f16-3266-4ab3-8153-a248b015228f'
ORDER BY created_at DESC
LIMIT 5;

-- 3. Now update your campaign to match server date (June 27)
UPDATE campaigns 
SET campaign_date = '2025-06-27',  -- Server thinks it's June 27
    scheduled_time = '00:00:00',    -- Past time
    status = 'pending',
    updated_at = NOW()
WHERE title = 'tsst send';

-- The worker should pick up the test message immediately
-- The campaign should trigger within 1 minute