-- Fix Issue 1: Add missing unique constraint for sequence_contacts
BEGIN;

-- Check existing constraints
SELECT conname, pg_get_constraintdef(oid) 
FROM pg_constraint 
WHERE conrelid = 'sequence_contacts'::regclass
AND contype = 'u';

-- Add the missing unique constraint that the ON CONFLICT needs
-- This should match what the Go code expects: (sequence_id, contact_phone, sequence_stepid)
ALTER TABLE sequence_contacts 
DROP CONSTRAINT IF EXISTS uq_sequence_contact_step;

ALTER TABLE sequence_contacts
ADD CONSTRAINT uq_sequence_contact_step 
UNIQUE (sequence_id, contact_phone, sequence_stepid);

-- Verify the constraint was created
SELECT conname, pg_get_constraintdef(oid) 
FROM pg_constraint 
WHERE conrelid = 'sequence_contacts'::regclass
AND contype = 'u';

COMMIT;