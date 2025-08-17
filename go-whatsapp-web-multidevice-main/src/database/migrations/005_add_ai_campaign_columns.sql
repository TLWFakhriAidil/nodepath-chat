-- Add AI campaign columns to campaigns table
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS ai VARCHAR(10);
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS "limit" INTEGER DEFAULT 0;

-- Create index for AI campaigns
CREATE INDEX IF NOT EXISTS idx_campaigns_ai ON campaigns(ai) WHERE ai IS NOT NULL;
