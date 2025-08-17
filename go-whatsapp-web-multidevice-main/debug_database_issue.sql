-- Diagnostic queries to check what's happening
-- Run these in your PostgreSQL to debug

-- 1. Check if we're in the right database
SELECT current_database();

-- 2. Check which schema we're using
SELECT current_schema();

-- 3. List all schemas
SELECT schema_name FROM information_schema.schemata;

-- 4. Check if leads table exists and where
SELECT schemaname, tablename 
FROM pg_tables 
WHERE tablename = 'leads';

-- 5. Check columns in leads table
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'leads' 
ORDER BY ordinal_position;

-- 6. Check if trigger column exists in leads
SELECT column_name 
FROM information_schema.columns 
WHERE table_name = 'leads' 
AND column_name = 'trigger';

-- 7. Check sequences table columns
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'sequences' 
AND column_name IN ('priority', 'is_active');

-- 8. Check sequence_contacts columns
SELECT column_name 
FROM information_schema.columns 
WHERE table_name = 'sequence_contacts' 
AND column_name IN ('current_trigger', 'processing_device_id', 'next_trigger_time');

-- 9. Try the failing query directly
SELECT l.trigger 
FROM leads l 
LIMIT 1;

-- 10. Check if there's a case sensitivity issue
SELECT column_name 
FROM information_schema.columns 
WHERE lower(table_name) = 'leads' 
AND lower(column_name) = 'trigger';