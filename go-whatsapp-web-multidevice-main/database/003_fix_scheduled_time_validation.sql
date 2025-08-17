-- Fix scheduled_time column to handle various formats properly

-- 1. First check the current state
SELECT 
    column_name,
    data_type,
    is_nullable
FROM information_schema.columns
WHERE table_name = 'campaigns'
AND column_name = 'scheduled_time';

-- 2. Backup current data
CREATE TABLE campaigns_time_backup AS
SELECT id, title, scheduled_time 
FROM campaigns;

-- 3. Convert scheduled_time to VARCHAR to handle flexible input
ALTER TABLE campaigns 
ALTER COLUMN scheduled_time TYPE VARCHAR(8);

-- 4. Clean up existing bad data
UPDATE campaigns 
SET scheduled_time = NULL
WHERE scheduled_time = ''
   OR scheduled_time = 'asdasd'
   OR scheduled_time !~ '^\d{2}:\d{2}(:\d{2})?$';

-- 5. Create a function to validate time format
CREATE OR REPLACE FUNCTION is_valid_time(time_str VARCHAR) 
RETURNS BOOLEAN AS $$
BEGIN
    IF time_str IS NULL OR time_str = '' THEN
        RETURN TRUE; -- NULL/empty is valid (means run immediately)
    END IF;
    
    -- Check if it matches HH:MM or HH:MM:SS format
    IF time_str ~ '^\d{2}:\d{2}(:\d{2})?$' THEN
        -- Try to cast it
        BEGIN
            PERFORM time_str::TIME;
            RETURN TRUE;
        EXCEPTION WHEN OTHERS THEN
            RETURN FALSE;
        END;
    END IF;
    
    RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

-- 6. Add a check constraint to prevent invalid times
ALTER TABLE campaigns 
ADD CONSTRAINT valid_scheduled_time 
CHECK (is_valid_time(scheduled_time));

-- 7. Fix your current campaign
UPDATE campaigns 
SET scheduled_time = '07:50:00',
    status = 'pending'
WHERE title = 'aqil';

-- 8. Verify the fix
SELECT 
    id,
    title,
    campaign_date,
    scheduled_time,
    status,
    CASE 
        WHEN scheduled_time IS NULL OR scheduled_time = '' THEN 'Run immediately'
        WHEN is_valid_time(scheduled_time) THEN 'Valid time: ' || scheduled_time
        ELSE 'INVALID TIME FORMAT'
    END as time_status
FROM campaigns
WHERE status = 'pending';