-- CRITICAL: Add unique constraints to prevent duplicates at database level
-- Run this SQL immediately!

-- First, remove any existing duplicates
-- For sequences
DELETE t1 FROM broadcast_messages t1
INNER JOIN broadcast_messages t2 
WHERE t1.id > t2.id
AND t1.sequence_stepid = t2.sequence_stepid 
AND t1.recipient_phone = t2.recipient_phone 
AND t1.device_id = t2.device_id
AND t1.sequence_stepid IS NOT NULL;

-- For campaigns
DELETE t1 FROM broadcast_messages t1
INNER JOIN broadcast_messages t2 
WHERE t1.id > t2.id
AND t1.campaign_id = t2.campaign_id 
AND t1.recipient_phone = t2.recipient_phone 
AND t1.device_id = t2.device_id
AND t1.campaign_id IS NOT NULL;

-- Add unique constraints
ALTER TABLE broadcast_messages 
ADD UNIQUE INDEX IF NOT EXISTS unique_sequence_message (
    sequence_stepid, 
    recipient_phone, 
    device_id
);

ALTER TABLE broadcast_messages 
ADD UNIQUE INDEX IF NOT EXISTS unique_campaign_message (
    campaign_id, 
    recipient_phone, 
    device_id
);

-- Verify constraints were added
SHOW INDEX FROM broadcast_messages WHERE Key_name IN ('unique_sequence_message', 'unique_campaign_message');
