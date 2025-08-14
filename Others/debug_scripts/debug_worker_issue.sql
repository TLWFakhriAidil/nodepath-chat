-- Debug worker issues
-- Check if there are any pending messages that should trigger workers

-- 1. Check if there are any pending broadcast messages
SELECT 
    COUNT(*) as pending_count,
    device_id,
    campaign_id,
    sequence_id,
    MIN(created_at) as oldest_pending,
    MAX(created_at) as newest_pending
FROM broadcast_messages
WHERE status = 'pending'
GROUP BY device_id, campaign_id, sequence_id;

-- 2. Check your connected device
SELECT 
    id,
    user_id,
    device_name,
    status,
    last_seen
FROM user_devices
WHERE device_name = 'aqil';

-- 3. Check if campaign has been processed
SELECT 
    c.id,
    c.title,
    c.status,
    c.campaign_date,
    c.scheduled_time,
    COUNT(bm.id) as message_count
FROM campaigns c
LEFT JOIN broadcast_messages bm ON bm.campaign_id = c.id
WHERE c.title = 'test'
GROUP BY c.id, c.title, c.status, c.campaign_date, c.scheduled_time;

-- 4. If no messages, manually create one to test worker
-- First update campaign to pending
UPDATE campaigns 
SET status = 'pending' 
WHERE title = 'test' AND status != 'sent';

-- 5. Check if you have any leads for the campaign
SELECT 
    l.*,
    ud.device_name,
    ud.status as device_status
FROM leads l
JOIN user_devices ud ON l.device_id = ud.id
WHERE l.niche LIKE '%VITAC%' 
AND l.target_status = 'customer'
LIMIT 5;