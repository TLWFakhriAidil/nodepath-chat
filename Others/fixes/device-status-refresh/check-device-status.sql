-- Fix for device status display issue
-- Run this SQL to check your device status

-- 1. Check current device status
SELECT 
    id,
    device_name,
    status,
    phone,
    jid,
    last_seen,
    created_at
FROM user_devices
ORDER BY created_at DESC;

-- 2. If you see NULL values in phone/jid columns, fix the query in the code
-- The issue might be that the Go code can't handle NULL values properly

-- 3. Force update a specific device to online (replace YOUR_DEVICE_ID)
-- UPDATE user_devices 
-- SET status = 'online', 
--     last_seen = CURRENT_TIMESTAMP
-- WHERE id = 'YOUR_DEVICE_ID';

-- 4. Check if there are any constraints or triggers affecting the status column
SELECT 
    conname AS constraint_name,
    contype AS constraint_type,
    pg_get_constraintdef(oid) AS definition
FROM pg_constraint
WHERE conrelid = 'user_devices'::regclass;
