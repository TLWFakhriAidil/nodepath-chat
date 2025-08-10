-- IMMEDIATE FIX for your campaign

-- 1. Fix the scheduled_time to trigger immediately
UPDATE campaigns 
SET scheduled_time = NULL,
    status = 'pending'
WHERE title = 'aqil';

-- 2. Alternative: Set to current time to trigger now
UPDATE campaigns 
SET scheduled_time = TO_CHAR(CURRENT_TIME, 'HH24:MI:SS'),
    status = 'pending'
WHERE title = 'aqil';

-- 3. Check the result
SELECT 
    id,
    title,
    campaign_date,
    scheduled_time,
    status,
    'Ready to send NOW' as message
FROM campaigns
WHERE title = 'aqil';

-- This will make your campaign trigger in the next minute!