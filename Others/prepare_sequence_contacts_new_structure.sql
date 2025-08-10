-- Clean up existing sequence_contacts data to prepare for new structure
-- WARNING: This will reset all sequence progress!

-- First, backup existing data
CREATE TABLE IF NOT EXISTS sequence_contacts_backup AS 
SELECT * FROM sequence_contacts;

-- Clear the existing data
TRUNCATE TABLE sequence_contacts;

-- Add the sequence_stepid column if it doesn't exist
ALTER TABLE sequence_contacts 
ADD COLUMN IF NOT EXISTS sequence_stepid UUID;

-- Create index for the new unique constraint
CREATE UNIQUE INDEX IF NOT EXISTS idx_sequence_contacts_unique 
ON sequence_contacts(sequence_id, contact_phone, sequence_stepid) 
WHERE sequence_stepid IS NOT NULL;

-- Drop the old unique constraint if it exists
ALTER TABLE sequence_contacts 
DROP CONSTRAINT IF EXISTS sequence_contacts_sequence_id_contact_phone_key;

-- Now the system will create individual records for each step when leads are enrolled
