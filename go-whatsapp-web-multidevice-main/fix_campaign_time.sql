-- Fix campaigns with empty scheduled_time
UPDATE campaigns 
SET scheduled_time = '09:00:00'::TIME 
WHERE scheduled_time IS NULL 
   OR scheduled_time::text = '';

-- Add default value for future campaigns
ALTER TABLE campaigns 
ALTER COLUMN scheduled_time 
SET DEFAULT '09:00:00'::TIME;

-- Verify the fix
SELECT id, title, scheduled_date, scheduled_time 
FROM campaigns 
WHERE scheduled_time IS NULL 
   OR scheduled_time::text = '';
