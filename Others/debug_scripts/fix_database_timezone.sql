-- Fix PostgreSQL timezone
-- Run this in your database

-- 1. Check current timezone
SHOW timezone;

-- 2. Set timezone for current session
SET timezone = 'Asia/Kuala_Lumpur';

-- 3. Set default timezone for all sessions (requires admin)
ALTER DATABASE your_database_name SET timezone TO 'Asia/Kuala_Lumpur';

-- 4. Verify the change
SELECT NOW() as current_time_malaysia;

-- 5. Update all existing campaigns to use local date
UPDATE campaigns 
SET campaign_date = (campaign_date::timestamp AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Kuala_Lumpur')::date
WHERE campaign_date > CURRENT_DATE;