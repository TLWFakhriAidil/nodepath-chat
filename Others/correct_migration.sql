-- Check existing columns first
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'campaigns';

-- Add target_status column if it doesn't exist (this is different from status)
ALTER TABLE campaigns 
ADD COLUMN IF NOT EXISTS target_status VARCHAR(50) DEFAULT 'prospect';

-- Same for sequences
ALTER TABLE sequences 
ADD COLUMN IF NOT EXISTS target_status VARCHAR(50) DEFAULT 'prospect';

-- Update any NULL values to default
UPDATE campaigns SET target_status = 'prospect' WHERE target_status IS NULL;
UPDATE sequences SET target_status = 'prospect' WHERE target_status IS NULL;

-- Verify the columns were added
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'campaigns' 
AND column_name IN ('status', 'target_status');
