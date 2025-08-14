-- IMMEDIATE FIX: Check and create test lead if needed

-- 1. Check if you have ANY leads with VITAC niche and customer status
SELECT COUNT(*) as matching_leads
FROM leads
WHERE niche LIKE '%VITAC%' 
AND target_status = 'customer';

-- 2. If count is 0, create a test lead
-- First, get your device ID
SELECT id, device_name, user_id, status
FROM user_devices
WHERE device_name = 'aqil' AND status = 'connected';

-- 3. Create a test lead (replace DEVICE_ID and USER_ID with actual values)
INSERT INTO leads (
    device_id,
    user_id,
    name,
    phone,
    niche,
    target_status,
    journey,
    status,
    created_at,
    updated_at
) VALUES (
    'YOUR_DEVICE_ID_HERE',  -- Replace with actual device ID
    'YOUR_USER_ID_HERE',    -- Replace with actual user ID
    'Test Customer',
    '60123456789',          -- Your test phone number
    'VITAC',
    'customer',             -- Must be 'customer' to match campaign
    'Test lead for campaign',
    'new',
    NOW(),
    NOW()
);

-- 4. Force the campaign to run by updating its scheduled time to past
UPDATE campaigns 
SET scheduled_time = '00:00:00',  -- Set to midnight so it's definitely past
    updated_at = NOW()
WHERE title = 'tsst send' 
AND status = 'pending';

-- 5. The campaign trigger should pick this up within 1 minute
-- Check Worker Status page after running these queries