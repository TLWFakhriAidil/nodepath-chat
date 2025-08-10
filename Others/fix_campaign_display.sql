-- Quick fix for campaign display issues
-- Run this after the main migration

-- 1. Check current campaign data
SELECT id, title, campaign_date, scheduled_time, status 
FROM campaigns 
ORDER BY campaign_date DESC 
LIMIT 10;

-- 2. Fix any campaigns with malformed dates
UPDATE campaigns 
SET campaign_date = DATE(campaign_date)
WHERE campaign_date LIKE '%T%';

-- 3. Ensure scheduled_time is properly set for campaigns with time data
UPDATE campaigns 
SET scheduled_time = CURRENT_TIMESTAMP 
WHERE scheduled_time IS NULL 
  AND campaign_date >= CURRENT_DATE;

-- 4. Create a debug view to help troubleshoot
CREATE OR REPLACE VIEW campaign_debug_view AS
SELECT 
    c.id,
    c.title,
    c.campaign_date,
    TO_CHAR(c.campaign_date, 'YYYY-MM-DD') as formatted_date,
    c.scheduled_time,
    TO_CHAR(c.scheduled_time, 'HH24:MI') as formatted_time,
    c.status,
    c.user_id
FROM campaigns c
ORDER BY c.campaign_date DESC;

-- 5. Check the results
SELECT * FROM campaign_debug_view;
