-- Run this SQL directly in your PostgreSQL database to fix the sequence steps issue immediately

-- 1. Add ALL missing columns that the Go application expects
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS trigger VARCHAR(255);
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS next_trigger VARCHAR(255);
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS trigger_delay_hours INTEGER DEFAULT 24;
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS is_entry_point BOOLEAN DEFAULT false;
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS image_url TEXT;
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS min_delay_seconds INTEGER DEFAULT 10;
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS max_delay_seconds INTEGER DEFAULT 30;

-- 2. Add missing columns to sequences table
ALTER TABLE sequences ADD COLUMN IF NOT EXISTS step_count INTEGER DEFAULT 0;
ALTER TABLE sequences ADD COLUMN IF NOT EXISTS total_steps INTEGER DEFAULT 0;

-- 3. Fix all NULL values in sequence_steps
UPDATE sequence_steps 
SET 
    trigger = COALESCE(trigger, 'step_' || id),
    next_trigger = COALESCE(next_trigger, ''),
    trigger_delay_hours = COALESCE(trigger_delay_hours, 24),
    is_entry_point = COALESCE(is_entry_point, false),
    min_delay_seconds = COALESCE(min_delay_seconds, 10),
    max_delay_seconds = COALESCE(max_delay_seconds, 30),
    message_type = COALESCE(message_type, 'text'),
    send_time = CASE 
        WHEN send_time = 'Invalid Date' OR send_time IS NULL OR send_time = '' 
        THEN '10:00' 
        ELSE send_time 
    END,
    time_schedule = CASE 
        WHEN time_schedule = 'Invalid Date' OR time_schedule IS NULL OR time_schedule = '' 
        THEN '10:00'
        ELSE time_schedule 
    END;

-- 4. Update sequence counts
UPDATE sequences 
SET 
    step_count = (SELECT COUNT(*) FROM sequence_steps WHERE sequence_id = sequences.id),
    total_steps = (SELECT COUNT(*) FROM sequence_steps WHERE sequence_id = sequences.id);

-- 5. Show results
SELECT 'Sequences with step counts:' as message;
SELECT id, name, step_count, total_steps FROM sequences WHERE step_count > 0;

SELECT 'Sample sequence steps:' as message;
SELECT id, sequence_id, day_number, trigger, content FROM sequence_steps LIMIT 10;
