-- Test script to verify auto-migration worked

-- 1. Check sequence_steps table structure
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'sequence_steps'
ORDER BY ordinal_position;

-- 2. Verify timestamp columns are removed
SELECT 
    CASE 
        WHEN EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'sequence_steps' AND column_name = 'send_time') THEN 'ERROR: send_time still exists!'
        WHEN EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'sequence_steps' AND column_name = 'created_at') THEN 'ERROR: created_at still exists!'
        WHEN EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'sequence_steps' AND column_name = 'updated_at') THEN 'ERROR: updated_at still exists!'
        WHEN EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'sequence_steps' AND column_name = 'day') THEN 'ERROR: day still exists!'
        WHEN EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'sequence_steps' AND column_name = 'schedule_time') THEN 'ERROR: schedule_time still exists!'
        ELSE 'SUCCESS: All timestamp columns removed!'
    END as migration_status;

-- 3. Verify new columns exist
SELECT 
    CASE 
        WHEN NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'sequence_steps' AND column_name = 'trigger') THEN 'ERROR: trigger column missing!'
        WHEN NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'sequence_steps' AND column_name = 'next_trigger') THEN 'ERROR: next_trigger column missing!'
        WHEN NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'sequence_steps' AND column_name = 'trigger_delay_hours') THEN 'ERROR: trigger_delay_hours column missing!'
        WHEN NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'sequence_steps' AND column_name = 'is_entry_point') THEN 'ERROR: is_entry_point column missing!'
        ELSE 'SUCCESS: All required columns exist!'
    END as new_columns_status;

-- 4. Check your data
SELECT * FROM sequence_steps 
WHERE sequence_id IN (
    '394d567f-e5bd-476d-ae7c-c39f74819d70',
    'b7319119-4d21-43cc-97e2-9644ded0608c'
);

-- 5. Check sequence step counts
SELECT 
    s.id,
    s.name,
    COUNT(ss.id) as actual_step_count
FROM sequences s
LEFT JOIN sequence_steps ss ON s.id = ss.sequence_id
WHERE s.user_id = 'de078f16-3266-4ab3-8153-a248b015228f'
GROUP BY s.id, s.name;
