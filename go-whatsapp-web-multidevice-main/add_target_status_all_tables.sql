-- Add target_status column to all three tables with default 'customer'

-- 1. Add to leads table
ALTER TABLE leads 
ADD COLUMN IF NOT EXISTS target_status TEXT DEFAULT 'customer';

-- 2. Add to campaigns table  
ALTER TABLE campaigns 
ADD COLUMN IF NOT EXISTS target_status TEXT DEFAULT 'customer';

-- 3. Add to sequences table
ALTER TABLE sequences 
ADD COLUMN IF NOT EXISTS target_status TEXT DEFAULT 'customer';

-- Update any NULL values to default
UPDATE leads SET target_status = 'customer' WHERE target_status IS NULL;
UPDATE campaigns SET target_status = 'customer' WHERE target_status IS NULL;
UPDATE sequences SET target_status = 'customer' WHERE target_status IS NULL;

-- Migrate existing status data to target_status for leads
UPDATE leads SET target_status = status WHERE target_status = 'customer' AND status IN ('prospect', 'customer');

-- Verify the columns were added
SELECT 
    table_name,
    column_name, 
    data_type,
    column_default
FROM information_schema.columns 
WHERE table_name IN ('leads', 'campaigns', 'sequences')
AND column_name = 'target_status'
ORDER BY table_name;
