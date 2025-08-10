-- Direct worker test - bypass campaign system

-- 1. Get your user and device info
SELECT 
    u.id as user_id,
    ud.id as device_id,
    ud.device_name,
    ud.status
FROM users u
JOIN user_devices ud ON ud.user_id = u.id
WHERE u.email = 'aqil@gmail.com'
AND ud.device_name = 'aqil';

-- 2. Create a simple broadcast message directly
-- Replace the IDs with actual values from query above
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
) 
SELECT 
    u.id,
    ud.id,
    '60108924904',  -- Your test phone
    'text',
    'Direct test message - if you receive this, workers are OK!',
    'pending',
    NOW(),
    NOW(),
    NOW()
FROM users u
JOIN user_devices ud ON ud.user_id = u.id
WHERE u.email = 'aqil@gmail.com'
AND ud.device_name = 'aqil'
AND ud.status = 'connected';

-- 3. Check if message was created
SELECT * FROM broadcast_messages 
WHERE content LIKE '%Direct test message%'
ORDER BY created_at DESC;

-- If this creates a message but worker doesn't pick it up,
-- then the broadcast manager might not be running properly