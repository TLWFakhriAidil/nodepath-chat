-- Debug query to test what Go application should be getting
-- Run this to see if the query is working in PostgreSQL

-- This is the exact query that Go application runs:
SELECT 
    id, sequence_id, day_number, 
    COALESCE(trigger, '') as trigger, 
    COALESCE(next_trigger, '') as next_trigger,
    COALESCE(trigger_delay_hours, 24) as trigger_delay_hours,
    COALESCE(is_entry_point, false) as is_entry_point,
    COALESCE(message_type, 'text') as message_type, 
    COALESCE(content, '') as content, 
    COALESCE(media_url, '') as media_url, 
    COALESCE(caption, '') as caption, 
    COALESCE(send_time, '') as send_time,
    COALESCE(time_schedule, '') as time_schedule
FROM sequence_steps
WHERE sequence_id = '394d567f-e5bd-476d-ae7c-c39f74819d70'
ORDER BY day_number ASC;

-- Also check if the columns exist:
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'sequence_steps' 
AND column_name IN (
    'trigger', 'next_trigger', 'trigger_delay_hours', 
    'is_entry_point', 'image_url', 'min_delay_seconds', 'max_delay_seconds'
)
ORDER BY column_name;
