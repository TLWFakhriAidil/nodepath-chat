-- Comprehensive fix for sequence_steps issues

-- 1. First, let's see what's actually in the database
SELECT 
    id,
    sequence_id,
    day_number,
    trigger,
    created_at,
    updated_at
FROM sequence_steps
WHERE sequence_id IN (
    '394d567f-e5bd-476d-ae7c-c39f74819d70',
    'b7319119-4d21-43cc-97e2-9644ded0608c'
);

-- 2. Fix any invalid timestamps
UPDATE sequence_steps
SET 
    created_at = NOW(),
    updated_at = NOW()
WHERE 
    created_at IS NULL 
    OR updated_at IS NULL
    OR created_at::text = 'Invalid Date'
    OR updated_at::text = 'Invalid Date';

-- 3. Ensure all required fields have values
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
    send_time = COALESCE(send_time, ''),
    time_schedule = COALESCE(time_schedule, '10:00')
WHERE sequence_id IN (
    '394d567f-e5bd-476d-ae7c-c39f74819d70',
    'b7319119-4d21-43cc-97e2-9644ded0608c'
);

-- 4. Verify the fix
SELECT 
    id,
    sequence_id,
    day_number,
    trigger,
    message_type,
    content,
    created_at,
    updated_at
FROM sequence_steps
WHERE sequence_id IN (
    '394d567f-e5bd-476d-ae7c-c39f74819d70',
    'b7319119-4d21-43cc-97e2-9644ded0608c'
);

-- 5. Test the exact query the Go app uses
SELECT 
    id, sequence_id, 
    COALESCE(day_number, 1) as day_number,
    COALESCE(trigger, '') as trigger, 
    COALESCE(next_trigger, '') as next_trigger,
    COALESCE(trigger_delay_hours, 24) as trigger_delay_hours,
    COALESCE(is_entry_point, false) as is_entry_point,
    COALESCE(message_type, 'text') as message_type, 
    COALESCE(content, '') as content, 
    COALESCE(media_url, '') as media_url, 
    COALESCE(caption, '') as caption, 
    COALESCE(send_time, '') as send_time,
    COALESCE(time_schedule, '') as time_schedule,
    created_at, updated_at
FROM sequence_steps
WHERE sequence_id = '394d567f-e5bd-476d-ae7c-c39f74819d70'
ORDER BY day_number ASC;
