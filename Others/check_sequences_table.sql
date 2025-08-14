-- Test query to check what columns exist and what data we have
SELECT 
    column_name,
    data_type,
    is_nullable
FROM information_schema.columns
WHERE table_name = 'sequences'
ORDER BY ordinal_position;

-- Also show sample data
SELECT id, name, niche, status, schedule_time, is_active 
FROM sequences 
LIMIT 5;
