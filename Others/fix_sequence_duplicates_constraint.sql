-- Fix for sequence duplicate messages
-- Add unique constraint to prevent duplicates at database level

-- First, remove existing duplicates (keep the oldest one)
DELETE bm1 FROM broadcast_messages bm1
INNER JOIN broadcast_messages bm2 
WHERE bm1.sequence_stepid = bm2.sequence_stepid 
AND bm1.recipient_phone = bm2.recipient_phone 
AND bm1.device_id = bm2.device_id
AND bm1.created_at > bm2.created_at;

-- Add unique index to prevent future duplicates
-- This will prevent duplicate messages at the database level
ALTER TABLE broadcast_messages 
ADD UNIQUE INDEX unique_sequence_message (
    sequence_stepid, 
    recipient_phone, 
    device_id
);

-- Note: This might fail if there are still duplicates
-- Run the DELETE query again if needed
