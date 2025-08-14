-- Debug query to check device assignment for your campaign
-- Run this in your PostgreSQL to verify

-- Check how many messages per device for the campaign
SELECT 
    bm.device_id,
    ud.device_name,
    COUNT(*) as message_count,
    STRING_AGG(DISTINCT bm.status, ', ') as statuses
FROM broadcast_messages bm
LEFT JOIN user_devices ud ON ud.id = bm.device_id
WHERE bm.campaign_id = [YOUR_CAMPAIGN_ID]  -- Replace with actual campaign ID
GROUP BY bm.device_id, ud.device_name
ORDER BY message_count DESC;

-- Check all messages with their device assignments
SELECT 
    bm.id,
    bm.recipient_phone,
    bm.device_id,
    ud.device_name,
    bm.status,
    bm.created_at
FROM broadcast_messages bm
LEFT JOIN user_devices ud ON ud.id = bm.device_id
WHERE bm.campaign_id = [YOUR_CAMPAIGN_ID]  -- Replace with actual campaign ID
ORDER BY bm.device_id, bm.created_at;

-- Check if there are any NULL device_ids
SELECT COUNT(*) as null_device_count
FROM broadcast_messages
WHERE campaign_id = [YOUR_CAMPAIGN_ID]  -- Replace with actual campaign ID
AND device_id IS NULL;
