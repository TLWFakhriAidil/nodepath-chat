-- Fix for duplicate sequence_contacts records
-- Run this SQL to clean up duplicates and add proper constraints

-- 1. First, let's see the duplicates
SELECT sequence_id, contact_phone, current_step, COUNT(*) as count
FROM sequence_contacts
GROUP BY sequence_id, contact_phone, current_step
HAVING COUNT(*) > 1;

-- 2. Delete duplicate records, keeping only one per contact/step
-- Keep the 'active' or 'completed' one if exists, otherwise keep the first 'pending'
WITH duplicates AS (
    SELECT id,
           ROW_NUMBER() OVER (
               PARTITION BY sequence_id, contact_phone, current_step 
               ORDER BY 
                   CASE status 
                       WHEN 'completed' THEN 1
                       WHEN 'active' THEN 2
                       WHEN 'pending' THEN 3
                       ELSE 4
                   END,
                   created_at ASC
           ) as rn
    FROM sequence_contacts
)
DELETE FROM sequence_contacts
WHERE id IN (
    SELECT id FROM duplicates WHERE rn > 1
);

-- 3. Add unique constraint to prevent future duplicates
-- Drop existing constraint if exists
ALTER TABLE sequence_contacts 
DROP CONSTRAINT IF EXISTS uk_sequence_contact_step;

-- Add new unique constraint
ALTER TABLE sequence_contacts
ADD CONSTRAINT uk_sequence_contact_step 
UNIQUE (sequence_id, contact_phone, current_step);

-- 4. Fix the ON CONFLICT issue in enrollContactInSequence
-- The error "no unique or exclusion constraint matching ON CONFLICT" means
-- the unique constraint doesn't match what the code expects
-- Let's check existing constraints
SELECT 
    tc.constraint_name,
    tc.constraint_type,
    kcu.column_name
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu 
    ON tc.constraint_name = kcu.constraint_name
WHERE tc.table_name = 'sequence_contacts'
ORDER BY tc.constraint_name, kcu.ordinal_position;

-- 5. Add the constraint that matches the code's ON CONFLICT clause
ALTER TABLE sequence_contacts
DROP CONSTRAINT IF EXISTS uk_sequence_contact_stepid;

ALTER TABLE sequence_contacts
ADD CONSTRAINT uk_sequence_contact_stepid
UNIQUE (sequence_id, contact_phone, sequence_stepid);

-- 6. Clean up any contacts in weird states
-- Reset stuck 'active' contacts that have been processing too long
UPDATE sequence_contacts
SET status = 'pending',
    processing_device_id = NULL,
    processing_started_at = NULL
WHERE status = 'active'
  AND processing_started_at < NOW() - INTERVAL '1 hour';

-- 7. Ensure only one active step per sequence/contact
-- Find cases where multiple steps are active
WITH multiple_active AS (
    SELECT sequence_id, contact_phone, COUNT(*) as active_count
    FROM sequence_contacts
    WHERE status = 'active'
    GROUP BY sequence_id, contact_phone
    HAVING COUNT(*) > 1
)
SELECT sc.*
FROM sequence_contacts sc
JOIN multiple_active ma 
  ON sc.sequence_id = ma.sequence_id 
  AND sc.contact_phone = ma.contact_phone
WHERE sc.status = 'active'
ORDER BY sc.contact_phone, sc.current_step;

-- 8. If multiple active found, keep only the lowest step active
WITH ranked_active AS (
    SELECT id,
           ROW_NUMBER() OVER (
               PARTITION BY sequence_id, contact_phone 
               ORDER BY current_step ASC
           ) as rn
    FROM sequence_contacts
    WHERE status = 'active'
)
UPDATE sequence_contacts
SET status = 'pending'
WHERE id IN (
    SELECT id FROM ranked_active WHERE rn > 1
);

-- 9. Verify the fix
SELECT 
    sequence_id,
    contact_phone,
    contact_name,
    current_step,
    status,
    next_trigger_time,
    completed_at
FROM sequence_contacts
WHERE sequence_id = '1-4ed6-891c-bcb7d12baa8a'
ORDER BY contact_phone, current_step;
