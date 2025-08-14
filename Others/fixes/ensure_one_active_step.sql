-- Critical fix to ensure only one active step per contact
BEGIN;

-- First, clean up any duplicate active steps
WITH ranked_active AS (
    SELECT id, 
           ROW_NUMBER() OVER (PARTITION BY sequence_id, contact_phone ORDER BY current_step) as rn
    FROM sequence_contacts
    WHERE status = 'active'
)
UPDATE sequence_contacts
SET status = 'pending'
WHERE id IN (
    SELECT id FROM ranked_active WHERE rn > 1
);

-- Add a partial unique index to prevent multiple active steps per contact
CREATE UNIQUE INDEX IF NOT EXISTS idx_one_active_per_contact 
ON sequence_contacts (sequence_id, contact_phone) 
WHERE status = 'active';

-- Add check constraint to ensure next_trigger_time is set
ALTER TABLE sequence_contacts 
ADD CONSTRAINT check_next_trigger_time_not_null 
CHECK (next_trigger_time IS NOT NULL);

COMMIT;
