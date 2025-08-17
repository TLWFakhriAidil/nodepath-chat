-- Check duplicate messages for specific phone number
SELECT 
    bm.id,
    bm.recipient_phone,
    bm.content,
    bm.status,
    bm.created_at,
    bm.sent_at,
    bm.device_id,
    ud.device_name,
    bm.campaign_id,
    bm.sequence_id,
    bm.processing_worker_id,
    bm.processing_started_at
FROM broadcast_messages bm
LEFT JOIN user_devices ud ON bm.device_id = ud.id
WHERE bm.recipient_phone LIKE '%128198574%' 
   OR bm.recipient_phone = '+60128198574'
   OR bm.recipient_phone = '60128198574'
ORDER BY bm.created_at DESC
LIMIT 20;

-- Check if there are exact duplicates
SELECT 
    recipient_phone,
    content,
    COUNT(*) as duplicate_count,
    GROUP_CONCAT(id) as message_ids,
    GROUP_CONCAT(status) as statuses,
    GROUP_CONCAT(device_id) as device_ids,
    GROUP_CONCAT(processing_worker_id) as worker_ids,
    MIN(created_at) as first_created,
    MAX(created_at) as last_created
FROM broadcast_messages
WHERE recipient_phone LIKE '%128198574%'
GROUP BY recipient_phone, content
HAVING COUNT(*) > 1
ORDER BY duplicate_count DESC;

-- Check sequence-specific duplicates
SELECT 
    sequence_id,
    sequence_stepid,
    recipient_phone,
    COUNT(*) as count
FROM broadcast_messages
WHERE recipient_phone LIKE '%128198574%'
  AND sequence_id IS NOT NULL
GROUP BY sequence_id, sequence_stepid, recipient_phone
HAVING COUNT(*) > 1;

-- Check campaign-specific duplicates  
SELECT 
    campaign_id,
    recipient_phone,
    COUNT(*) as count
FROM broadcast_messages
WHERE recipient_phone LIKE '%128198574%'
  AND campaign_id IS NOT NULL
GROUP BY campaign_id, recipient_phone
HAVING COUNT(*) > 1;