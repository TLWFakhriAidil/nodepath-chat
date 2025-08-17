-- Fix scheduled_time column type issue

-- 1. Check current column type
SELECT 
    column_name,
    data_type,
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_name = 'campaigns'
AND column_name = 'scheduled_time';

-- 2. Backup current data
SELECT id, title, scheduled_time 
FROM campaigns 
WHERE scheduled_time IS NOT NULL;

-- 3. Option A: Convert to VARCHAR (recommended for flexibility)
ALTER TABLE campaigns 
ALTER COLUMN scheduled_time TYPE VARCHAR(8) USING scheduled_time::text;

-- 4. Option B: Convert to TIME type
-- ALTER TABLE campaigns 
-- ALTER COLUMN scheduled_time TYPE TIME USING scheduled_time::time;

-- 5. Option C: Drop the column entirely (if you don't need scheduled times)
-- ALTER TABLE campaigns DROP COLUMN scheduled_time;

-- 6. Update existing campaigns to have proper format
UPDATE campaigns 
SET scheduled_time = '00:00:00'
WHERE scheduled_time IS NOT NULL 
AND LENGTH(scheduled_time::text) > 8;

-- 7. Set your campaign to trigger immediately
UPDATE campaigns 
SET scheduled_time = '',  -- Empty string = trigger immediately
    status = 'pending'
WHERE title = 'tsst send';

-- 8. Verify the changes
SELECT 
    id,
    title,
    campaign_date,
    scheduled_time,
    status
FROM campaigns
WHERE status = 'pending';