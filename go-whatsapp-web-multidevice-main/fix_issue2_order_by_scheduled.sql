-- Fix for Issue 2: Order messages by scheduled_at
-- This ensures messages are sent in the correct sequence order

-- Current problematic ORDER BY:
-- ORDER BY bm.group_id, bm.group_order, bm.created_at ASC

-- Fixed ORDER BY:
-- ORDER BY bm.scheduled_at ASC, bm.group_id, bm.group_order

-- The complete fixed query for GetPendingMessages should be:
/*
SELECT bm.id, bm.user_id, bm.device_id, bm.campaign_id, bm.sequence_id, 
    bm.recipient_phone, bm.recipient_name, bm.message_type, bm.content AS message, bm.media_url, 
    bm.scheduled_at, bm.group_id, bm.group_order,
    COALESCE(c.min_delay_seconds, s.min_delay_seconds, 10) AS min_delay,
    COALESCE(c.max_delay_seconds, s.max_delay_seconds, 30) AS max_delay
FROM broadcast_messages bm
LEFT JOIN campaigns c ON bm.campaign_id = c.id
LEFT JOIN sequences s ON bm.sequence_id = s.id
WHERE bm.device_id = ? AND bm.status = 'pending'
AND (bm.scheduled_at IS NULL OR bm.scheduled_at <= ?)
ORDER BY bm.scheduled_at ASC, bm.group_id, bm.group_order
LIMIT ?
*/
