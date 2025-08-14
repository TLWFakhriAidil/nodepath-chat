-- Add target_status to campaigns table
ALTER TABLE campaigns 
ADD COLUMN IF NOT EXISTS target_status VARCHAR(50) DEFAULT 'all';

-- Add target_status to sequences table
ALTER TABLE sequences 
ADD COLUMN IF NOT EXISTS target_status VARCHAR(50) DEFAULT 'all';

-- Update existing campaigns/sequences to have 'all' as target_status
UPDATE campaigns SET target_status = 'all' WHERE target_status IS NULL;
UPDATE sequences SET target_status = 'all' WHERE target_status IS NULL;
