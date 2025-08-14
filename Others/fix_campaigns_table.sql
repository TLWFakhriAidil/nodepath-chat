-- Fix campaigns table to ensure all required columns exist
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS device_id UUID;

-- Check if any campaigns exist
SELECT id, user_id, campaign_date, title, niche, status FROM campaigns;

-- Show the table structure
SELECT column_name, data_type, is_nullable 
FROM information_schema.columns 
WHERE table_name = 'campaigns' 
ORDER BY ordinal_position;
