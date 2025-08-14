-- Fix campaigns table - add missing columns
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS niche VARCHAR(255);
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS image_url TEXT;
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS scheduled_time TIME;
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'scheduled';

-- Ensure all columns allow NULL where appropriate
ALTER TABLE campaigns ALTER COLUMN niche DROP NOT NULL;
ALTER TABLE campaigns ALTER COLUMN image_url DROP NOT NULL;
ALTER TABLE campaigns ALTER COLUMN scheduled_time DROP NOT NULL;

-- Add missing indexes if they don't exist
CREATE INDEX IF NOT EXISTS idx_campaigns_status ON campaigns(status);
CREATE INDEX IF NOT EXISTS idx_campaigns_scheduled ON campaigns(campaign_date, scheduled_time);

-- Show current table structure
SELECT column_name, data_type, is_nullable
FROM information_schema.columns
WHERE table_name = 'campaigns'
ORDER BY ordinal_position;
