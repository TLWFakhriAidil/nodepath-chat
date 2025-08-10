-- Fix timezone issue - change campaign date to match server date

-- 1. Check what date the server thinks it is
SELECT 
    NOW() as server_time_utc,
    NOW() AT TIME ZONE 'Asia/Kuala_Lumpur' as malaysia_time,
    NOW()::date as server_date,
    (NOW() AT TIME ZONE 'Asia/Kuala_Lumpur')::date as malaysia_date;

-- 2. Update your campaign to use server date (June 27)
UPDATE campaigns 
SET campaign_date = '2025-06-27',  -- Use server date
    scheduled_time = '00:00:00',   -- Set to past time
    status = 'pending',
    updated_at = NOW()
WHERE title = 'tsst send';

-- 3. Verify the update
SELECT 
    id,
    title,
    campaign_date,
    scheduled_time,
    status,
    NOW()::date as server_today,
    CASE 
        WHEN campaign_date = NOW()::date THEN 'MATCHES_SERVER_DATE'
        ELSE 'DATE_MISMATCH'
    END as date_check
FROM campaigns
WHERE title = 'tsst send';

-- The campaign should trigger within 1 minute after this update!