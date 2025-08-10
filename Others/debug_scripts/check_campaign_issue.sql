-- Debug why campaign didn't trigger
-- Check timezone and campaign status

-- 1. Check server time vs campaign time
SELECT 
    NOW() as server_time,
    NOW() AT TIME ZONE 'Asia/Kuala_Lumpur' as malaysia_time;

-- 2. Check your campaign details with time comparison
SELECT 
    id,
    title,
    campaign_date,
    scheduled_time,
    status,
    created_at,
    updated_at,
    -- Time calculations
    campaign_date || ' ' || scheduled_time as scheduled_datetime,
    NOW()::date as today,
    NOW()::time as current_time,
    CASE 
        WHEN campaign_date = NOW()::date AND scheduled_time::time < NOW()::time THEN 'Should have triggered'
        WHEN campaign_date < NOW()::date THEN 'Past date'
        WHEN campaign_date > NOW()::date THEN 'Future date'
        ELSE 'Not yet time'
    END as trigger_status
FROM campaigns
WHERE title = 'test'
ORDER BY created_at DESC;

-- 3. Check if campaign trigger service is running
SELECT * FROM logs 
WHERE created_at > NOW() - INTERVAL '10 minutes'
AND (message LIKE '%campaign%' OR message LIKE '%trigger%')
ORDER BY created_at DESC;

-- 4. Manual trigger check - see what would happen
SELECT 
    c.id as campaign_id,
    c.title,
    c.niche,
    c.target_status,
    COUNT(l.id) as matching_leads
FROM campaigns c
LEFT JOIN leads l ON l.niche LIKE '%' || c.niche || '%' 
    AND (c.target_status = 'all' OR l.target_status = c.target_status)
WHERE c.title = 'test'
GROUP BY c.id, c.title, c.niche, c.target_status;