-- Remove unique constraint on campaigns table to allow multiple campaigns per date
ALTER TABLE campaigns DROP CONSTRAINT IF EXISTS campaigns_user_id_campaign_date_key;

-- Add index for performance instead
CREATE INDEX IF NOT EXISTS idx_campaigns_user_date ON campaigns(user_id, campaign_date);