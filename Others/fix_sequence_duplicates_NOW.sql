-- IMMEDIATE FIX FOR SEQUENCE DUPLICATE MESSAGES
-- Run this SQL on your database NOW to prevent duplicates

-- Step 1: Check how many duplicates exist
SELECT 
    sequence_stepid,
    recipient_phone,
    device_id,
    COUNT(*) as duplicate_count,
    GROUP_CONCAT(id) as message_ids,
    GROUP_CONCAT(status) as statuses,
    GROUP_CONCAT(sent_at) as sent_times
FROM broadcast_messages 
WHERE sequence_stepid IS NOT NULL
GROUP BY sequence_stepid, recipient_phone, device_id
HAVING COUNT(*) > 1
ORDER BY COUNT(*) DESC
LIMIT 20;

-- Step 2: Remove existing duplicates (keep the oldest one)
DELETE bm1 FROM broadcast_messages bm1
INNER JOIN broadcast_messages bm2 
WHERE bm1.sequence_stepid = bm2.sequence_stepid 
AND bm1.recipient_phone = bm2.recipient_phone 
AND bm1.device_id = bm2.device_id
AND bm1.created_at > bm2.created_at;

-- Step 3: Add unique constraint to prevent future duplicates
-- This will prevent the database from accepting duplicate messages
ALTER TABLE broadcast_messages 
ADD UNIQUE INDEX unique_sequence_message (
    sequence_stepid, 
    recipient_phone, 
    device_id
);

-- Note: If step 3 fails with "Duplicate entry" error, run step 2 again
-- until all duplicates are removed

-- Step 4: Verify the constraint was added
SHOW INDEX FROM broadcast_messages WHERE Key_name = 'unique_sequence_message';

-- Step 5: Test that duplicates are prevented
-- This should fail with a duplicate key error (which is what we want)
-- INSERT INTO broadcast_messages (id, sequence_stepid, recipient_phone, device_id, status)
-- VALUES ('test-id-1', 'test-step', 'test-phone', 'test-device', 'pending'),
--        ('test-id-2', 'test-step', 'test-phone', 'test-device', 'pending');
