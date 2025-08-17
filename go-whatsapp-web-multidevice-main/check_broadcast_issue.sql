-- Check what's happening with broadcast_messages

-- 1. See all broadcast_messages for sequences
SELECT 
    bm.id,
    bm.recipient_phone,
    bm.device_id,
    bm.sequence_id,
    bm.status,
    bm.created_at,
    bm.sent_at,
    LEFT(bm.content, 50) as message_preview
FROM broadcast_messages bm
WHERE bm.sequence_id IS NOT NULL
ORDER BY bm.created_at DESC
LIMIT 20;

-- 2. Check sequence_contacts status vs broadcast_messages
SELECT 
    sc.contact_phone,
    sc.current_step,
    sc.status as sequence_status,
    sc.current_trigger,
    sc.assigned_device_id,
    COUNT(bm.id) as broadcast_count,
    STRING_AGG(DISTINCT bm.status, ', ') as broadcast_statuses
FROM sequence_contacts sc
LEFT JOIN broadcast_messages bm 
    ON bm.recipient_phone = sc.contact_phone 
    AND bm.sequence_id = sc.sequence_id
GROUP BY sc.contact_phone, sc.current_step, sc.status, sc.current_trigger, sc.assigned_device_id
ORDER BY sc.contact_phone, sc.current_step;

-- 3. Check if there's a trigger creating multiple broadcast messages
SELECT 
    COUNT(*) as total_broadcast_messages,
    COUNT(DISTINCT recipient_phone) as unique_recipients,
    COUNT(DISTINCT sequence_id) as unique_sequences,
    MIN(created_at) as first_created,
    MAX(created_at) as last_created
FROM broadcast_messages
WHERE sequence_id IS NOT NULL;
