-- WhatsApp Campaign System Database Migration
-- Run this script to update your existing database
-- Date: June 27, 2025

-- 1. Backup existing scheduled_time data before converting
CREATE TABLE IF NOT EXISTS campaigns_backup AS 
SELECT id, scheduled_time::text as scheduled_time_text 
FROM campaigns 
WHERE scheduled_time IS NOT NULL;

-- 2. Fix campaigns table
ALTER TABLE campaigns ALTER COLUMN scheduled_time TYPE TIMESTAMP USING 
  CASE 
    WHEN scheduled_time IS NOT NULL THEN 
      (campaign_date::date + scheduled_time::time)::timestamp
    ELSE NULL
  END;

-- 3. Add missing columns to campaigns
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS min_delay_seconds INTEGER DEFAULT 10;
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS max_delay_seconds INTEGER DEFAULT 30;

-- 4. Add missing columns to sequences
ALTER TABLE sequences ADD COLUMN IF NOT EXISTS schedule_time VARCHAR(10);
ALTER TABLE sequences ADD COLUMN IF NOT EXISTS min_delay_seconds INTEGER DEFAULT 10;
ALTER TABLE sequences ADD COLUMN IF NOT EXISTS max_delay_seconds INTEGER DEFAULT 30;

-- 5. Add missing columns to sequence_steps
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS schedule_time VARCHAR(10);

-- 6. Create helpful view for campaign calendar
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

-- 7. Update any NULL delay values to defaults
UPDATE campaigns SET min_delay_seconds = 10 WHERE min_delay_seconds IS NULL;
UPDATE campaigns SET max_delay_seconds = 30 WHERE max_delay_seconds IS NULL;
UPDATE sequences SET min_delay_seconds = 10 WHERE min_delay_seconds IS NULL;
UPDATE sequences SET max_delay_seconds = 30 WHERE max_delay_seconds IS NULL;
UPDATE sequence_steps SET min_delay_seconds = 10 WHERE min_delay_seconds IS NULL;
UPDATE sequence_steps SET max_delay_seconds = 30 WHERE max_delay_seconds IS NULL;

-- 8. Add indexes for better performance
CREATE INDEX IF NOT EXISTS idx_campaigns_scheduled_time ON campaigns(scheduled_time);
CREATE INDEX IF NOT EXISTS idx_sequences_schedule_time ON sequences(schedule_time);

-- 9. Verify the migration
SELECT 
    'Campaigns with scheduled time' as check_type,
    COUNT(*) as count
FROM campaigns 
WHERE scheduled_time IS NOT NULL
UNION ALL
SELECT 
    'Sequences with schedule time' as check_type,
    COUNT(*) as count
FROM sequences 
WHERE schedule_time IS NOT NULL;

-- Note: After running this migration, ensure your application is updated to handle the new TIMESTAMP type
