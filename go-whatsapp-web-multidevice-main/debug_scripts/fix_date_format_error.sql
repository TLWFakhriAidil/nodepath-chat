-- FIX DATE FORMAT ERROR

-- 1. Check current campaign date format
SELECT 
    id,
    title,
    campaign_date,
    scheduled_time,
    status
FROM campaigns
WHERE title = 'tsst send';

-- 2. Fix the date format (remove timestamp, keep only date)
UPDATE campaigns 
SET campaign_date = '2025-06-28',  -- Just the date, no timestamp
    scheduled_time = '01:10:00',   -- Just the time
    status = 'pending'
WHERE title = 'tsst send';

-- 3. Verify the fix
SELECT 
    id,
    title,
    campaign_date,
    scheduled_time,
    campaign_date || ' ' || scheduled_time || ':00' as combined_datetime
FROM campaigns
WHERE title = 'tsst send';

-- The campaign should trigger in the next minute!