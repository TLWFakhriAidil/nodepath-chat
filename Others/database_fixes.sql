-- Database fixes for WhatsApp Campaign System
-- Date: June 27, 2025

-- 1. Fix campaigns table - change scheduled_time to TIMESTAMP for better date/time handling
ALTER TABLE campaigns DROP COLUMN IF EXISTS scheduled_time;
ALTER TABLE campaigns ADD COLUMN scheduled_time TIMESTAMP;

-- 2. Add missing min_delay and max_delay columns to campaigns table
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS min_delay_seconds INTEGER DEFAULT 10;
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS max_delay_seconds INTEGER DEFAULT 30;

-- 3. Fix sequences table - add schedule_time column
ALTER TABLE sequences ADD COLUMN IF NOT EXISTS schedule_time VARCHAR(10);
ALTER TABLE sequences ADD COLUMN IF NOT EXISTS min_delay_seconds INTEGER DEFAULT 10;
ALTER TABLE sequences ADD COLUMN IF NOT EXISTS max_delay_seconds INTEGER DEFAULT 30;

-- 4. Update sequence_steps table to ensure all columns exist
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS schedule_time VARCHAR(10);
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS min_delay_seconds INTEGER DEFAULT 10;
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS max_delay_seconds INTEGER DEFAULT 30;

-- 5. Create a view to help with campaign display
CREATE OR REPLACE VIEW campaign_calendar_view AS
SELECT 
    c.id,
    c.user_id,
    c.title,
    c.niche,
    c.message,
    c.image_url,
    c.campaign_date,
    c.scheduled_time,
    c.status,
    c.min_delay_seconds,
    c.max_delay_seconds,
    TO_CHAR(c.scheduled_time, 'HH24:MI') as display_time,
    TO_CHAR(c.campaign_date, 'YYYY-MM-DD') as calendar_date
FROM campaigns c
ORDER BY c.campaign_date, c.scheduled_time;

-- 6. Update any existing NULL scheduled_time to current time
UPDATE campaigns 
SET scheduled_time = CURRENT_TIMESTAMP 
WHERE scheduled_time IS NULL;

-- 7. Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_campaigns_calendar ON campaigns(campaign_date, scheduled_time);
CREATE INDEX IF NOT EXISTS idx_sequences_schedule ON sequences(schedule_time);
