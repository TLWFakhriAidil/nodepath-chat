-- Debug why campaign isn't finding the lead

-- 1. Check the exact campaign criteria
SELECT 
    id,
    title,
    niche,
    target_status,
    status,
    campaign_date,
    scheduled_time,
    user_id
FROM campaigns
WHERE title LIKE '%tsst%' OR title LIKE '%test%'
ORDER BY created_at DESC;

-- 2. Check the exact lead data
SELECT 
    l.id,
    l.name,
    l.phone,
    l.niche,
    l.target_status,
    l.device_id,
    l.user_id,
    ud.device_name,
    ud.status as device_status
FROM leads l
JOIN user_devices ud ON l.device_id = ud.id
WHERE l.phone = '60108924904';

-- 3. Test the exact matching query the campaign uses
SELECT 
    l.*,
    c.title as campaign_title,
    c.niche as campaign_niche,
    c.target_status as campaign_target_status
FROM leads l
CROSS JOIN campaigns c
WHERE c.title LIKE '%tsst%'
AND l.niche LIKE '%' || c.niche || '%'
AND (c.target_status = 'all' OR l.target_status = c.target_status);

-- 4. Check if broadcast messages were created
SELECT 
    bm.id,
    bm.status,
    bm.recipient_phone,
    bm.created_at,
    c.title as campaign_title
FROM broadcast_messages bm
LEFT JOIN campaigns c ON bm.campaign_id = c.id
WHERE c.title LIKE '%tsst%' OR c.title LIKE '%test%'
ORDER BY bm.created_at DESC
LIMIT 10;

-- 5. Check the campaign trigger logs
SELECT * FROM logs 
WHERE created_at > NOW() - INTERVAL '30 minutes'
AND (message LIKE '%campaign%' OR message LIKE '%trigger%' OR message LIKE '%tsst%')
ORDER BY created_at DESC
LIMIT 20;