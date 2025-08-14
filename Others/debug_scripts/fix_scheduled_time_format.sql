-- Fix scheduled_time format issue

-- 1. Check the actual scheduled_time value
SELECT 
    id,
    title,
    campaign_date,
    scheduled_time,
    scheduled_time::text as time_as_text,
    LENGTH(scheduled_time::text) as time_length
FROM campaigns
WHERE title = 'tsst send';

-- 2. Update scheduled_time to proper format (HH:MM:SS)
UPDATE campaigns 
SET scheduled_time = '01:10:00',  -- Proper time format
    status = 'pending'
WHERE title = 'tsst send';

-- 3. Alternative: Set to past time to trigger immediately
UPDATE campaigns 
SET scheduled_time = '00:00:00',  -- Midnight - will trigger immediately
    status = 'pending'
WHERE title = 'tsst send';

-- 4. Check if there's a data type issue
SELECT 
    column_name,
    data_type,
    character_maximum_length
FROM information_schema.columns
WHERE table_name = 'campaigns'
AND column_name IN ('campaign_date', 'scheduled_time');