-- OPTIMIZED: Database migration for proper timezone handling
-- Based on PostgreSQL best practices

-- 1. Add TIMESTAMPTZ column for proper timezone support
ALTER TABLE campaigns 
ADD COLUMN IF NOT EXISTS scheduled_at TIMESTAMPTZ;

-- 2. Migrate existing data to TIMESTAMPTZ (assuming Malaysia timezone)
UPDATE campaigns 
SET scheduled_at = 
    CASE 
        WHEN scheduled_time IS NULL OR scheduled_time = '' THEN 
            campaign_date::TIMESTAMP AT TIME ZONE 'Asia/Kuala_Lumpur'
        ELSE 
            (campaign_date || ' ' || scheduled_time)::TIMESTAMP AT TIME ZONE 'Asia/Kuala_Lumpur'
    END
WHERE scheduled_at IS NULL;

-- 3. Create optimized indexes
CREATE INDEX IF NOT EXISTS idx_campaigns_scheduled_at ON campaigns(scheduled_at) 
WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS idx_campaigns_scheduled_at_status ON campaigns(scheduled_at, status);

-- 4. Set session timezone for consistent querying
SET TIMEZONE = 'Asia/Kuala_Lumpur';

-- 5. Test the migration - view campaigns in different timezones
SELECT 
    id,
    title,
    campaign_date,
    scheduled_time,
    scheduled_at,
    scheduled_at AT TIME ZONE 'Asia/Kuala_Lumpur' as malaysia_time,
    scheduled_at AT TIME ZONE 'UTC' as utc_time,
    CASE 
        WHEN scheduled_at <= CURRENT_TIMESTAMP THEN 'Ready to send'
        ELSE 'Scheduled in ' || (scheduled_at - CURRENT_TIMESTAMP)
    END as status_check
FROM campaigns
WHERE status = 'pending'
ORDER BY scheduled_at;

-- 6. Function to get campaigns ready to send (timezone-aware)
CREATE OR REPLACE FUNCTION get_pending_campaigns()
RETURNS TABLE (
    id INTEGER,
    title VARCHAR,
    scheduled_at TIMESTAMPTZ,
    local_time TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        c.id,
        c.title,
        c.scheduled_at,
        c.scheduled_at AT TIME ZONE 'Asia/Kuala_Lumpur' as local_time
    FROM campaigns c
    WHERE c.status = 'pending'
    AND c.scheduled_at <= CURRENT_TIMESTAMP
    ORDER BY c.scheduled_at;
END;
$$ LANGUAGE plpgsql;