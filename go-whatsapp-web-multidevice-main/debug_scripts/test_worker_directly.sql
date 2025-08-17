-- Test if workers are functioning

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
    'de078f16-3266-4ab3-8153-a248b015228f',
    '2de48db2-f1ab-4d81-8a26-58b01df75bdf',
    '60108924904',
    'text',
    'WORKER TEST: If you get this, workers are OK! Time: ' || NOW()::text,
    'pending',
    NOW(),
    NOW(),
    NOW()
);

-- 2. Check if worker picks it up
SELECT 
    id,
    status,
    content,
    created_at,
    updated_at
FROM broadcast_messages
WHERE recipient_phone = '60108924904'
ORDER BY created_at DESC
LIMIT 5;