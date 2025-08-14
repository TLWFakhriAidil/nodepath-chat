-- Fix sequence_steps table by removing timestamp columns and ensuring data integrity

-- 1. First, check what columns exist
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'sequence_steps'
ORDER BY ordinal_position;

-- 2. Remove problematic timestamp columns
ALTER TABLE sequence_steps 
DROP COLUMN IF EXISTS send_time CASCADE,
DROP COLUMN IF EXISTS created_at CASCADE,
DROP COLUMN IF EXISTS updated_at CASCADE,
DROP COLUMN IF EXISTS day CASCADE,
DROP COLUMN IF EXISTS schedule_time CASCADE;

-- 3. Add missing columns if they don't exist
ALTER TABLE sequence_steps 
ADD COLUMN IF NOT EXISTS message_type VARCHAR(50) DEFAULT 'text',
ADD COLUMN IF NOT EXISTS trigger VARCHAR(255) DEFAULT '',
ADD COLUMN IF NOT EXISTS next_trigger VARCHAR(255) DEFAULT '',
ADD COLUMN IF NOT EXISTS trigger_delay_hours INTEGER DEFAULT 24,
ADD COLUMN IF NOT EXISTS is_entry_point BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS min_delay_seconds INTEGER DEFAULT 10,
ADD COLUMN IF NOT EXISTS max_delay_seconds INTEGER DEFAULT 30,
ADD COLUMN IF NOT EXISTS delay_days INTEGER DEFAULT 0;

-- 4. Update existing data to ensure proper values
UPDATE sequence_steps
SET 
    day_number = COALESCE(day_number, 1),
    message_type = COALESCE(message_type, 'text'),
    trigger = COALESCE(trigger, ''),
    next_trigger = COALESCE(next_trigger, ''),
    trigger_delay_hours = COALESCE(trigger_delay_hours, 24),
    is_entry_point = COALESCE(is_entry_point, false),
    content = COALESCE(content, ''),
    media_url = COALESCE(media_url, ''),
    caption = COALESCE(caption, ''),
    time_schedule = COALESCE(time_schedule, '10:00'),
    min_delay_seconds = COALESCE(min_delay_seconds, 10),
    max_delay_seconds = COALESCE(max_delay_seconds, 30),
    delay_days = COALESCE(delay_days, 0);

-- 5. Verify the final structure
SELECT column_name, data_type, column_default
FROM information_schema.columns 
WHERE table_name = 'sequence_steps'
ORDER BY ordinal_position;

-- 6. Check your existing data
SELECT * FROM sequence_steps 
WHERE sequence_id IN (
    '394d567f-e5bd-476d-ae7c-c39f74819d70',
    'b7319119-4d21-43cc-97e2-9644ded0608c'
);

-- 7. Update sequences table to reflect actual step counts
UPDATE sequences s
SET step_count = (
    SELECT COUNT(*) 
    FROM sequence_steps ss 
    WHERE ss.sequence_id = s.id
)
WHERE s.user_id = 'de078f16-3266-4ab3-8153-a248b015228f';

-- 8. Verify sequences now show correct counts
SELECT 
    s.id,
    s.name,
    s.step_count,
    (SELECT COUNT(*) FROM sequence_steps WHERE sequence_id = s.id) as actual_steps
FROM sequences s
WHERE s.user_id = 'de078f16-3266-4ab3-8153-a248b015228f';
