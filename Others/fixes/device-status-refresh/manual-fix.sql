-- Manual fix for device showing offline
-- Run these queries in your PostgreSQL database

-- 1. First, check the current status of your device
SELECT id, device_name, status, phone, jid, last_seen
FROM user_devices
WHERE id = '8ccc6409-124f-4f68-b618-0e64e69d61b8';

-- 2. If it shows status='offline' and phone/jid are empty, 
--    manually update it to match what the logs show:
UPDATE user_devices
SET 
    status = 'online',
    phone = '60146674397',
    jid = '60146674397:52@s.whatsapp.net',
    last_seen = CURRENT_TIMESTAMP
WHERE id = '8ccc6409-124f-4f68-b618-0e64e69d61b8';

-- 3. Verify the update worked
SELECT id, device_name, status, phone, jid, last_seen
FROM user_devices
WHERE id = '8ccc6409-124f-4f68-b618-0e64e69d61b8';

-- 4. After running these queries, refresh your browser
--    The device should now show as "Connected" with phone number

-- Note: The issue is that the UpdateDeviceStatus call in the Go code
-- is not finding the correct device ID from the session, so the update
-- never happens. This manual update fixes it temporarily.
