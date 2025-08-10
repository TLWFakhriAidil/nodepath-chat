-- Comprehensive Database Migration for Timezone Support
-- Run this to fix all timezone issues

-- 1. Set session timezone
SET TIMEZONE = 'Asia/Kuala_Lumpur';

-- 2. Add TIMESTAMPTZ columns if they don't exist
ALTER TABLE campaigns 
ADD COLUMN IF NOT EXISTS scheduled_at TIMESTAMPTZ;

ALTER TABLE sequences 
ADD COLUMN IF NOT EXISTS created_at_tz TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE sequence_contacts
ADD COLUMN IF NOT EXISTS last_message_at_tz TIMESTAMPTZ;

-- 3. Migrate existing campaign data to TIMESTAMPTZ
UPDATE campaigns 
SET scheduled_at = 
    CASE 
        WHEN scheduled_time IS NULL OR scheduled_time = '' OR scheduled_time = '00:00:00' THEN 
            campaign_date::TIMESTAMP AT TIME ZONE 'Asia/Kuala_Lumpur'
        ELSE 
            (campaign_date || ' ' || scheduled_time)::TIMESTAMP AT TIME ZONE 'Asia/Kuala_Lumpur'
    END
WHERE scheduled_at IS NULL;

-- 4. Fix your current pending campaign
UPDATE campaigns 
SET status = 'pending',
    scheduled_at = CURRENT_TIMESTAMP - INTERVAL '1 minute'
WHERE title IN ('tsst send', 'amasd')
AND status != 'sent';

-- 5. Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_campaigns_scheduled_at ON campaigns(scheduled_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_campaigns_status ON campaigns(status);
CREATE INDEX IF NOT EXISTS idx_sequence_contacts_active ON sequence_contacts(status) WHERE status = 'active';

-- 6. Create a view for easy campaign monitoring
CREATE OR REPLACE VIEW campaign_status_view AS
SELECT 
    id,
    title,
    campaign_date,
    scheduled_time,
    scheduled_at,
    scheduled_at AT TIME ZONE 'Asia/Kuala_Lumpur' as malaysia_time,
    status,
    CASE 
        WHEN status = 'sent' THEN 'Already sent'
        WHEN status != 'pending' THEN 'Not pending'
        WHEN scheduled_at IS NULL AND (scheduled_time IS NULL OR scheduled_time = '' OR scheduled_time = '00:00:00') THEN 'Ready - No time set'
        WHEN scheduled_at <= CURRENT_TIMESTAMP THEN 'Ready - Time passed'
        WHEN COALESCE(scheduled_at, (campaign_date || ' ' || COALESCE(scheduled_time, '00:00:00'))::TIMESTAMP AT TIME ZONE 'Asia/Kuala_Lumpur') <= CURRENT_TIMESTAMP THEN 'Ready - Calculated time passed'
        ELSE 'Waiting - Scheduled for ' || to_char(scheduled_at AT TIME ZONE 'Asia/Kuala_Lumpur', 'DD Mon YYYY HH24:MI')
    END as execution_status,
    created_at,
    updated_at
FROM campaigns
ORDER BY COALESCE(scheduled_at, created_at) DESC;

-- 7. Check campaign readiness
SELECT * FROM campaign_status_view WHERE status = 'pending';

-- 8. Force immediate execution for testing
-- UPDATE campaigns SET scheduled_at = CURRENT_TIMESTAMP - INTERVAL '1 hour' WHERE title = 'YOUR_CAMPAIGN_NAME' AND status = 'pending';