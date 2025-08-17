-- Check duplicate messages for phone number +60128198574 for today (August 10, 2025)
-- Checking all possible formats of the phone number

-- First, let's see all messages for this phone number today
SELECT 
    id,
    recipient_phone,
    LEFT(content, 100) as message_preview,
    status,
    created_at,
    sent_at,
    device_id,
    campaign_id,
    sequence_id,
    processing_worker_id
FROM broadcast_messages
WHERE (recipient_phone = '+60128198574' 
    OR recipient_phone = '60128198574'
    OR recipient_phone = '0128198574'
    OR recipient_phone LIKE '%128198574%')
    AND DATE(created_at) = '2025-08-10'
ORDER BY created_at DESC;

-- Count duplicates by content for today
SELECT 
    recipient_phone,
    LEFT(content, 100) as message_preview,
    COUNT(*) as duplicate_count,
    MIN(created_at) as first_created,
    MAX(created_at) as last_created,
    GROUP_CONCAT(status) as all_statuses,
    GROUP_CONCAT(DISTINCT device_id) as devices_used
FROM broadcast_messages
WHERE (recipient_phone = '+60128198574' 
    OR recipient_phone = '60128198574'
    OR recipient_phone = '0128198574'
    OR recipient_phone LIKE '%128198574%')
    AND DATE(created_at) = '2025-08-10'
GROUP BY recipient_phone, content
HAVING COUNT(*) > 1;

-- Check exact time around 1:38 PM (13:38)
SELECT 
    id,
    recipient_phone,
    LEFT(content, 50) as message_preview,
    status,
    created_at,
    sent_at,
    TIME(created_at) as created_time,
    TIME(sent_at) as sent_time,
    processing_worker_id,
    device_id
FROM broadcast_messages
WHERE (recipient_phone = '+60128198574' 
    OR recipient_phone = '60128198574'
    OR recipient_phone = '0128198574'
    OR recipient_phone LIKE '%128198574%')
    AND DATE(created_at) = '2025-08-10'
    AND HOUR(created_at) BETWEEN 13 AND 14
ORDER BY created_at;
