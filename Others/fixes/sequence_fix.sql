-- Fix for sequence processing issues
-- This script will:
-- 1. Delete all existing sequence_contacts data
-- 2. Update sequence status to inactive (already done by user)
-- 3. Verify the fix

-- Step 1: Delete all sequence_contacts data
DELETE FROM sequence_contacts;

-- Step 2: Verify deletion
SELECT COUNT(*) as remaining_records FROM sequence_contacts;

-- Step 3: Check sequence status
SELECT id, name, status, is_active FROM sequences;

-- Step 4: Create proper indexes if not exists
CREATE INDEX IF NOT EXISTS idx_sequence_contacts_status_time 
ON sequence_contacts(status, next_trigger_time) 
WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_sequence_contacts_phone_seq 
ON sequence_contacts(contact_phone, sequence_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_sequence_contacts_unique_step
ON sequence_contacts(sequence_id, contact_phone, sequence_stepid) 
WHERE sequence_stepid IS NOT NULL;

-- Step 5: Add check constraint to ensure current_step is reasonable
ALTER TABLE sequence_contacts 
DROP CONSTRAINT IF EXISTS check_current_step_range;

ALTER TABLE sequence_contacts 
ADD CONSTRAINT check_current_step_range 
CHECK (current_step >= 1 AND current_step <= 100);
