-- Add unique constraint to prevent duplicate message creation
-- This ensures only ONE message per recipient + sequence step

-- First, clean up existing duplicates (keep the first one)
DELETE bm1 FROM broadcast_messages bm1
INNER JOIN broadcast_messages bm2 
WHERE bm1.recipient_phone = bm2.recipient_phone
  AND bm1.sequence_stepid = bm2.sequence_stepid
  AND bm1.sequence_stepid IS NOT NULL
  AND bm1.status = 'pending'
  AND bm2.status = 'pending'
  AND bm1.created_at > bm2.created_at;

-- Add unique index to prevent future duplicates
-- This will prevent the same recipient from being enrolled in the same step twice
ALTER TABLE broadcast_messages 
ADD UNIQUE INDEX unique_sequence_enrollment (recipient_phone, sequence_stepid);
