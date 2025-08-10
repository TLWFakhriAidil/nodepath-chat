-- Check and fix campaigns table

-- 1. First, let's see the current structure
SELECT column_name, data_type, is_nullable, column_default
FROM information_schema.columns
WHERE table_name = 'campaigns'
ORDER BY ordinal_position;

-- 2. Add scheduled_time column if it doesn't exist
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS scheduled_time TIME;

-- 3. Add any other missing columns
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS device_id UUID;
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS niche VARCHAR(255);

-- 4. Check the data
SELECT id, campaign_date, title, scheduled_time, status, niche
FROM campaigns
ORDER BY campaign_date DESC;
