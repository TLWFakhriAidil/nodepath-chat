-- IMMEDIATE FIX: Handle empty scheduled_time values

-- 1. Check what's in scheduled_time for your campaigns
SELECT 
    id,
    title,
    campaign_date,
    scheduled_time,
    CASE 
        WHEN scheduled_time IS NULL THEN 'NULL'
        WHEN scheduled_time = '' THEN 'EMPTY STRING'
        ELSE scheduled_time
    END as time_check,
    status
FROM campaigns
WHERE status = 'pending';

-- 2. Fix empty string scheduled_time values
UPDATE campaigns 
SET scheduled_time = NULL
WHERE scheduled_time = ''
AND status = 'pending';

-- 3. Alternative: Set to a valid time
UPDATE campaigns 
SET scheduled_time = '00:00:00'
WHERE (scheduled_time = '' OR scheduled_time IS NULL)
AND status = 'pending';

-- 4. For your specific campaign, set to trigger immediately
UPDATE campaigns 
SET scheduled_time = NULL,
    status = 'pending'
WHERE title = 'aqil';

-- 5. Verify the fix
SELECT 
    id,
    title,
    campaign_date,
    scheduled_time,
    status
FROM campaigns
WHERE title = 'aqil';