-- Quick fix: Create a test campaign for today's server date

-- 1. First, find out what date the server thinks is "today"
SELECT 
    NOW()::date as server_today,
    NOW()::time as current_time;

-- 2. Create a new test campaign with server's date
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
)
SELECT 
    user_id,
    'test_now_' || NOW()::time,
    'VITAC',
    'customer',
    'Test message: Hello from campaign test!',
    NOW()::date,  -- Use server's current date
    '00:00:00',   -- Past time
    10,
    30,
    'pending',
    NOW(),
    NOW()
FROM campaigns
WHERE title = 'tsst send'
LIMIT 1;

-- 3. The campaign trigger should pick this up immediately
-- Check the logs after running this