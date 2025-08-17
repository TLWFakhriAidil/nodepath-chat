-- Migration to change scheduled_time to time_schedule in campaigns and sequences
-- This provides a unified approach across both tables

-- 1. Add new time_schedule column to campaigns
ALTER TABLE campaigns 
ADD COLUMN IF NOT EXISTS time_schedule TEXT;

-- 2. Migrate existing data from scheduled_time to time_schedule
UPDATE campaigns 
SET time_schedule = scheduled_time
WHERE time_schedule IS NULL AND scheduled_time IS NOT NULL;

-- 3. Add new time_schedule column to sequences (rename from schedule_time)
ALTER TABLE sequences 
ADD COLUMN IF NOT EXISTS time_schedule TEXT;

-- 4. Migrate existing data from schedule_time to time_schedule in sequences
UPDATE sequences 
SET time_schedule = schedule_time
WHERE time_schedule IS NULL AND schedule_time IS NOT NULL;

-- 5. Create validation function for time_schedule
CREATE OR REPLACE FUNCTION is_valid_time_schedule(time_str TEXT) 
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

-- 6. Add check constraints for time_schedule
ALTER TABLE campaigns 
ADD CONSTRAINT valid_campaign_time_schedule 
CHECK (is_valid_time_schedule(time_schedule));

ALTER TABLE sequences 
ADD CONSTRAINT valid_sequence_time_schedule 
CHECK (is_valid_time_schedule(time_schedule));

-- 7. Update the campaign trigger function to use time_schedule
CREATE OR REPLACE FUNCTION get_pending_campaigns_with_time_schedule()
RETURNS TABLE (
    id INTEGER,
    title VARCHAR,
    campaign_date DATE,
    time_schedule TEXT,
    scheduled_at TIMESTAMPTZ,
    local_time TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        c.id,
        c.title,
        c.campaign_date,
        c.time_schedule,
        CASE 
            WHEN c.time_schedule IS NULL OR c.time_schedule = '' THEN 
                c.campaign_date::TIMESTAMP AT TIME ZONE 'Asia/Kuala_Lumpur'
            ELSE 
                (c.campaign_date || ' ' || c.time_schedule)::TIMESTAMP AT TIME ZONE 'Asia/Kuala_Lumpur'
        END as scheduled_at,
        CASE 
            WHEN c.time_schedule IS NULL OR c.time_schedule = '' THEN 
                c.campaign_date::TIMESTAMP AT TIME ZONE 'Asia/Kuala_Lumpur'
            ELSE 
                (c.campaign_date || ' ' || c.time_schedule)::TIMESTAMP AT TIME ZONE 'Asia/Kuala_Lumpur'
        END AT TIME ZONE 'Asia/Kuala_Lumpur' as local_time
    FROM campaigns c
    WHERE c.status = 'pending'
    AND CASE 
            WHEN c.time_schedule IS NULL OR c.time_schedule = '' THEN 
                c.campaign_date::TIMESTAMP AT TIME ZONE 'Asia/Kuala_Lumpur'
            ELSE 
                (c.campaign_date || ' ' || c.time_schedule)::TIMESTAMP AT TIME ZONE 'Asia/Kuala_Lumpur'
        END <= CURRENT_TIMESTAMP
    ORDER BY scheduled_at;
END;
$$ LANGUAGE plpgsql;

-- 8. Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_campaigns_time_schedule ON campaigns(time_schedule) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_sequences_time_schedule ON sequences(time_schedule);

-- 9. Drop old columns (commented out for safety - run manually after verification)
-- ALTER TABLE campaigns DROP COLUMN IF EXISTS scheduled_time;
-- ALTER TABLE sequences DROP COLUMN IF EXISTS schedule_time;

-- 10. Verify migration
SELECT 
    'Campaigns' as table_name,
    COUNT(*) as total_records,
    COUNT(time_schedule) as records_with_time
FROM campaigns
UNION ALL
SELECT 
    'Sequences' as table_name,
    COUNT(*) as total_records,
    COUNT(time_schedule) as records_with_time
FROM sequences;
