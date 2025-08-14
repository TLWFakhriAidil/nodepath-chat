-- Simple check to see why steps aren't showing

-- 1. Check if your step exists
SELECT * FROM sequence_steps 
WHERE sequence_id = '394d567f-e5bd-476d-ae7c-c39f74819d70';

-- 2. Count the steps
SELECT sequence_id, COUNT(*) as step_count 
FROM sequence_steps 
WHERE sequence_id IN (
    '394d567f-e5bd-476d-ae7c-c39f74819d70',
    'b7319119-4d21-43cc-97e2-9644ded0608c'
)
GROUP BY sequence_id;

-- 3. Test the exact Go query - this is what the app is running
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
