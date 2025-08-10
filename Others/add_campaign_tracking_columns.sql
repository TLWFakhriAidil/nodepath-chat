-- Add columns to track campaign processing state
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS first_processed_at TIMESTAMP;
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS last_processed_at TIMESTAMP;
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS active_workers INTEGER DEFAULT 0;

-- Add indexes for better performance
CREATE INDEX IF NOT EXISTS idx_broadcast_messages_campaign_status ON broadcast_messages(campaign_id, status);
CREATE INDEX IF NOT EXISTS idx_campaigns_status_processed ON campaigns(status, first_processed_at);