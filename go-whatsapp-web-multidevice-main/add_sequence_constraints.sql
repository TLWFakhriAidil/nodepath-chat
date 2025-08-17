-- Add unique constraint to sequence_contacts table
-- This prevents duplicate enrollments for the same contact in the same sequence step

-- First, check if the constraint already exists
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'unique_sequence_contact_step'
    ) THEN
        -- Add the unique constraint
        ALTER TABLE sequence_contacts 
        ADD CONSTRAINT unique_sequence_contact_step 
        UNIQUE (sequence_id, contact_phone, sequence_stepid);
        
        RAISE NOTICE 'Unique constraint added successfully';
    ELSE
        RAISE NOTICE 'Unique constraint already exists';
    END IF;
END $$;

-- Also add an index for better performance on the pending steps query
CREATE INDEX IF NOT EXISTS idx_sequence_contacts_pending 
ON sequence_contacts(sequence_id, contact_phone, current_step, next_trigger_time) 
WHERE status = 'pending';

-- Add index for active steps query
CREATE INDEX IF NOT EXISTS idx_sequence_contacts_active
ON sequence_contacts(status, next_trigger_time)
WHERE status = 'active';

-- Verify the constraint was created
SELECT 
    conname as constraint_name,
    contype as constraint_type,
    array_to_string(conkey::int[], ', ') as constrained_columns
FROM pg_constraint 
WHERE conrelid = 'sequence_contacts'::regclass
AND conname = 'unique_sequence_contact_step';
