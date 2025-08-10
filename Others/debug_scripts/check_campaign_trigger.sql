-- Check campaign trigger readiness
-- Run this in your PostgreSQL to debug

-- 1. Check your campaign details
SELECT 
    id,
    user_id,
    title,
    niche,
    target_status,
    campaign_date,
    scheduled_time,
    status,
    created_at
FROM campaigns
WHERE title = 'test'
ORDER BY created_at DESC
LIMIT 1;

-- 2. Check if there are leads matching your campaign
SELECT COUNT(*) as matching_leads
FROM leads
WHERE niche LIKE '%VITAC%'
AND target_status = 'customer';

-- 3. Check connected devices
SELECT 
    id,
    user_id,
    device_name,
    status,
    last_seen
FROM user_devices
WHERE status = 'connected';

-- 4. Check if any broadcast messages were created
SELECT 
    COUNT(*) as message_count,
    status,
    MIN(created_at) as first_created,
    MAX(created_at) as last_created
FROM broadcast_messages
WHERE campaign_id = (SELECT id FROM campaigns WHERE title = 'test' ORDER BY created_at DESC LIMIT 1)
GROUP BY status;

-- 5. Check campaign trigger logs (if any)
SELECT * FROM logs 
WHERE message LIKE '%Campaign%test%' 
OR message LIKE '%Executing campaign%'
ORDER BY created_at DESC 
LIMIT 10;
