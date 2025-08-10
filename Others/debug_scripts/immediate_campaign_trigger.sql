-- IMMEDIATE FIX: Make your campaign trigger NOW

-- 1. Update the campaign to trigger immediately
UPDATE campaigns 
SET status = 'pending',
    scheduled_time = NULL,
    updated_at = NOW()
WHERE title = 'amasd';

-- 2. If you have the new scheduled_at column, use it
UPDATE campaigns 
SET scheduled_at = CURRENT_TIMESTAMP - INTERVAL '1 minute',
    status = 'pending'
WHERE title = 'amasd';

-- 3. Alternative: Set to past time in Malaysia timezone
UPDATE campaigns 
SET scheduled_at = (NOW() AT TIME ZONE 'Asia/Kuala_Lumpur' - INTERVAL '1 hour')::TIMESTAMPTZ,
    status = 'pending'
WHERE title = 'amasd';

-- 4. Verify the campaign is ready
SELECT 
    id,
    title,
    campaign_date,
    scheduled_time,
    scheduled_at,
    scheduled_at AT TIME ZONE 'Asia/Kuala_Lumpur' as malaysia_time,
    CASE 
        WHEN COALESCE(scheduled_at, (campaign_date || ' ' || COALESCE(scheduled_time, '00:00:00'))::TIMESTAMP AT TIME ZONE 'Asia/Kuala_Lumpur') <= CURRENT_TIMESTAMP 
        THEN 'READY TO SEND'
        ELSE 'NOT YET TIME'
    END as status
FROM campaigns
WHERE title = 'amasd';