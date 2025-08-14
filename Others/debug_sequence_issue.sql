-- Debug script to find out why sequence steps aren't showing
-- Run this in your PostgreSQL database

-- 1. Check if sequence_steps table has the required columns
SELECT 
    column_name, 
    data_type, 
    is_nullable,
    column_default
FROM information_schema.columns 
WHERE table_name = 'sequence_steps'
ORDER BY ordinal_position;

-- 2. Check if there are any steps for your sequences
SELECT 
    s.id as sequence_id,
    s.name as sequence_name,
    COUNT(ss.id) as step_count
FROM sequences s
LEFT JOIN sequence_steps ss ON s.id = ss.sequence_id
WHERE s.user_id = 'de078f16-3266-4ab3-8153-a248b015228f'
GROUP BY s.id, s.name;

-- 3. Show the actual step data
SELECT * FROM sequence_steps 
WHERE sequence_id IN (
    '394d567f-e5bd-476d-ae7c-c39f74819d70',
    'b7319119-4d21-43cc-97e2-9644ded0608c'
);

-- 4. Try the exact query that Go is using
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

-- 5. If the above query fails, let's check which columns are missing
DO $$
DECLARE
    missing_columns text[];
    required_columns text[] := ARRAY[
        'id', 'sequence_id', 'day_number', 'trigger', 'next_trigger',
        'trigger_delay_hours', 'is_entry_point', 'message_type', 
        'content', 'media_url', 'caption', 'send_time', 
        'time_schedule', 'created_at', 'updated_at'
    ];
    col text;
BEGIN
    missing_columns := ARRAY[]::text[];
    
    FOREACH col IN ARRAY required_columns
    LOOP
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'sequence_steps' AND column_name = col
        ) THEN
            missing_columns := array_append(missing_columns, col);
        END IF;
    END LOOP;
    
    IF array_length(missing_columns, 1) > 0 THEN
        RAISE NOTICE 'Missing columns: %', missing_columns;
    ELSE
        RAISE NOTICE 'All required columns exist';
    END IF;
END $$;
