-- EMERGENCY FIX: Re-link devices to users after migration
-- This script fixes the issue where user_id was set to NULL during CHAR(36) conversion

-- Step 1: Check current state
SELECT 'Devices with NULL user_id:' as status, COUNT(*) as count 
FROM device_setting_nodepath 
WHERE user_id IS NULL;

SELECT 'Total devices:' as status, COUNT(*) as count 
FROM device_setting_nodepath;

SELECT 'Total users:' as status, COUNT(*) as count 
FROM users_nodepath;

-- Step 2: If you have only ONE user, link ALL devices to that user
-- UNCOMMENT AND RUN THIS if you have a single user:
/*
UPDATE device_setting_nodepath d
SET d.user_id = (SELECT id FROM users_nodepath LIMIT 1)
WHERE d.user_id IS NULL;
*/

-- Step 3: Verify the fix
SELECT 'Devices with NULL user_id after fix:' as status, COUNT(*) as count 
FROM device_setting_nodepath 
WHERE user_id IS NULL;

SELECT 'Devices properly linked:' as status, COUNT(*) as count 
FROM device_setting_nodepath 
WHERE user_id IS NOT NULL;

-- Step 4: Check your specific user's devices
-- Replace 'YOUR_USER_ID' with your actual user ID
/*
SELECT d.id_device, d.device_name, d.user_id, u.username
FROM device_setting_nodepath d
LEFT JOIN users_nodepath u ON d.user_id = u.id
WHERE u.username = 'YOUR_USERNAME';
*/
