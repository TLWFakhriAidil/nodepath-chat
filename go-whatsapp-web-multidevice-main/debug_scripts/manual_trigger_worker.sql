-- Manual test to trigger worker
-- Run these queries step by step

-- 1. First, get your user ID and device ID
SELECT 
    u.id as user_id,
    u.email,
    ud.id as device_id,
    ud.device_name,
    ud.status as device_status,
    ud.phone
FROM users u
JOIN user_devices ud ON ud.user_id = u.id
WHERE ud.device_name = 'aqil';

-- 2. Get campaign ID
SELECT id, title, status, niche, target_status 
FROM campaigns 
WHERE title = 'test';

-- 3. Create a test broadcast message manually
-- Replace the values with actual IDs from above queries
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
) VALUES (
    'YOUR_USER_ID_HERE',  -- Replace with actual user ID
    'YOUR_DEVICE_ID_HERE', -- Replace with actual device ID
    YOUR_CAMPAIGN_ID_HERE, -- Replace with actual campaign ID (number, no quotes)
    '60123456789',        -- Test phone number
    'text',
    'Test message to trigger worker',
    'pending',
    NOW(),
    NOW(),
    NOW()
);

-- 4. Check if message was created
SELECT * FROM broadcast_messages 
WHERE status = 'pending' 
ORDER BY created_at DESC 
LIMIT 5;

-- 5. The worker should pick this up within seconds
-- Check worker status in the UI after running this