-- Add target_status to campaigns table if it doesn't exist
ALTER TABLE campaigns 
ADD COLUMN IF NOT EXISTS target_status VARCHAR(50) DEFAULT 'prospect';

-- Add target_status to sequences table if it doesn't exist
ALTER TABLE sequences 
ADD COLUMN IF NOT EXISTS target_status VARCHAR(50) DEFAULT 'prospect';

-- Update any existing NULL values
UPDATE campaigns SET target_status = 'prospect' WHERE target_status IS NULL;
UPDATE sequences SET target_status = 'prospect' WHERE target_status IS NULL;

-- Remove 'all' option if it exists
UPDATE campaigns SET target_status = 'prospect' WHERE target_status = 'all';
UPDATE sequences SET target_status = 'prospect' WHERE target_status = 'all';
