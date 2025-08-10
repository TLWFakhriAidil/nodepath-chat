-- Fix sequence_steps data to ensure all columns have proper values
-- This script helps fix any null or invalid data in your sequence_steps table

-- 1. First, let's see what data we have
SELECT 
    id,
    sequence_id,
    day_number,
    trigger,
    next_trigger,
    trigger_delay_hours,
    is_entry_point,
    message_type,
    content,
    created_at
FROM sequence_steps
ORDER BY sequence_id, day_number;

-- 2. Update any NULL values to proper defaults
UPDATE sequence_steps 
SET 
    day_number = COALESCE(day_number, 1),
    trigger = COALESCE(trigger, ''),
    next_trigger = COALESCE(next_trigger, ''),
    trigger_delay_hours = COALESCE(trigger_delay_hours, 24),
    is_entry_point = COALESCE(is_entry_point, false),
    message_type = COALESCE(message_type, 'text'),
    content = COALESCE(content, ''),
    media_url = COALESCE(media_url, ''),
    caption = COALESCE(caption, ''),
    send_time = COALESCE(send_time, ''),
    time_schedule = COALESCE(time_schedule, '10:00')
WHERE 
    day_number IS NULL 
    OR trigger IS NULL
    OR next_trigger IS NULL
    OR trigger_delay_hours IS NULL
    OR is_entry_point IS NULL
    OR message_type IS NULL;

-- 3. For your specific row, ensure it has proper values
UPDATE sequence_steps 
SET 
    day_number = 1,
    trigger = 'asdaxx',
    next_trigger = 'asdaxx_day2',
    trigger_delay_hours = 24,
    is_entry_point = true,
    message_type = 'text',
    time_schedule = '10:00'
WHERE id = '14cd1695-ee3a-4e9d-86cb-6729c2483aae';

-- 4. Verify the update
SELECT 
    id,
    sequence_id,
    day_number,
    trigger,
    next_trigger,
    trigger_delay_hours,
    is_entry_point,
    message_type,
    content,
    time_schedule
FROM sequence_steps
WHERE sequence_id IN (
    '394d567f-e5bd-476d-ae7c-c39f74819d70',
    'b7319119-4d21-43cc-97e2-9644ded0608c'
);

-- 5. Update the step count in sequences table
UPDATE sequences s
SET step_count = (
    SELECT COUNT(*) 
    FROM sequence_steps ss 
    WHERE ss.sequence_id = s.id
)
WHERE s.id IN (
    '394d567f-e5bd-476d-ae7c-c39f74819d70',
    'b7319119-4d21-43cc-97e2-9644ded0608c'
);

-- 6. Verify sequences now show correct step counts
SELECT 
    id,
    name,
    step_count,
    (SELECT COUNT(*) FROM sequence_steps WHERE sequence_id = s.id) as actual_steps
FROM sequences s
WHERE user_id = 'de078f16-3266-4ab3-8153-a248b015228f';
