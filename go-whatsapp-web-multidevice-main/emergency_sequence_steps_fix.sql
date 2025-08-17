-- 🚨 EMERGENCY FIX: Sequence Steps Missing Columns Issue
-- This SQL will fix the root cause of the sequence steps not being retrieved

-- PROBLEM: The Go application expects certain columns in sequence_steps that might not exist
-- SOLUTION: Add all missing columns and fix data integrity

-- 1. Add ALL missing columns that the Go query expects
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS trigger VARCHAR(255);
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS next_trigger VARCHAR(255);
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS trigger_delay_hours INTEGER DEFAULT 24;
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS is_entry_point BOOLEAN DEFAULT false;
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS image_url TEXT;
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS min_delay_seconds INTEGER DEFAULT 10;
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS max_delay_seconds INTEGER DEFAULT 30;

-- Add missing columns to sequences table for counts
ALTER TABLE sequences ADD COLUMN IF NOT EXISTS total_steps INTEGER DEFAULT 0;
ALTER TABLE sequences ADD COLUMN IF NOT EXISTS contact_count INTEGER DEFAULT 0;
ALTER TABLE sequences ADD COLUMN IF NOT EXISTS contacts_count INTEGER DEFAULT 0;
ALTER TABLE sequences ADD COLUMN IF NOT EXISTS step_count INTEGER DEFAULT 0;

-- 2. Fix all existing steps with proper values
UPDATE sequence_steps 
SET 
    trigger = CASE 
        WHEN trigger IS NULL OR trigger = '' THEN 
            CASE 
                WHEN sequence_id = '394d567f-e5bd-476d-ae7c-c39f74819d70' THEN 'asdaxx_day1'
                WHEN sequence_id = 'b7319119-4d21-43cc-97e2-9644ded0608c' THEN 'FUNP_day1'
                ELSE CONCAT('trigger_day', COALESCE(day_number, 1))
            END
        ELSE trigger 
    END,
    next_trigger = CASE 
        WHEN next_trigger IS NULL OR next_trigger = '' THEN 
            CASE 
                WHEN sequence_id = '394d567f-e5bd-476d-ae7c-c39f74819d70' THEN 'asdaxx_day2'
                WHEN sequence_id = 'b7319119-4d21-43cc-97e2-9644ded0608c' THEN 'FUNP_day2'
                ELSE CONCAT('trigger_day', COALESCE(day_number, 1) + 1)
            END
        ELSE next_trigger 
    END,
    trigger_delay_hours = COALESCE(trigger_delay_hours, 24),
    is_entry_point = COALESCE(is_entry_point, CASE WHEN day_number = 1 THEN true ELSE false END),
    min_delay_seconds = COALESCE(min_delay_seconds, 10),
    max_delay_seconds = COALESCE(max_delay_seconds, 30),
    send_time = CASE 
        WHEN send_time = 'Invalid Date' OR send_time IS NULL OR send_time = '' 
        THEN '10:00' 
        ELSE send_time 
    END,
    time_schedule = CASE 
        WHEN time_schedule = 'Invalid Date' OR time_schedule IS NULL OR time_schedule = '' 
        THEN COALESCE(send_time, '10:00')
        ELSE time_schedule 
    END,
    message_type = COALESCE(message_type, 'text'),
    day = COALESCE(day, day_number, 1),
    day_number = COALESCE(day_number, day, 1);

-- 3. Update sequence step counts to reflect actual data
UPDATE sequences 
SET 
    step_count = (SELECT COUNT(*) FROM sequence_steps WHERE sequence_id = sequences.id),
    total_steps = (SELECT COUNT(*) FROM sequence_steps WHERE sequence_id = sequences.id)
WHERE EXISTS (SELECT 1 FROM sequence_steps WHERE sequence_id = sequences.id);

-- 4. Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_sequence_steps_complete ON sequence_steps(sequence_id, day_number);
CREATE INDEX IF NOT EXISTS idx_sequence_steps_trigger_lookup ON sequence_steps(trigger) WHERE trigger IS NOT NULL;

-- 5. Show the results
SELECT 'Fixed Sequences:' as status;
SELECT id, name, step_count, total_steps FROM sequences ORDER BY created_at DESC LIMIT 10;

SELECT 'Fixed Steps:' as status;
SELECT id, sequence_id, day_number, trigger, content FROM sequence_steps ORDER BY sequence_id, day_number LIMIT 20;
