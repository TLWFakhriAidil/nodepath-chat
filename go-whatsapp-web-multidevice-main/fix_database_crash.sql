-- This SQL script will help identify and fix database issues after reverting code

-- 1. Check if problematic columns exist
SELECT 
    table_name,
    column_name,
    data_type,
    is_nullable
FROM 
    information_schema.columns
WHERE 
    table_schema = 'public'
    AND table_name IN ('leads_ai', 'ai_campaign_progress', 'campaigns', 'sequences', 'sequence_steps', 'broadcast_messages')
    AND column_name IN ('ai', 'limit', 'progress_percentage', 'total_contacts', 'active_contacts')
ORDER BY 
    table_name, 
    ordinal_position;

-- 2. Check if whatsmeow_message_secrets has the key column
SELECT 
    column_name,
    data_type
FROM 
    information_schema.columns
WHERE 
    table_name = 'whatsmeow_message_secrets';

-- 3. If you need to remove AI-related columns (CAREFUL - this removes data!)
-- Uncomment these lines only if you're sure:
-- ALTER TABLE campaigns DROP COLUMN IF EXISTS ai;
-- ALTER TABLE campaigns DROP COLUMN IF EXISTS "limit";
-- DROP TABLE IF EXISTS leads_ai CASCADE;
-- DROP TABLE IF EXISTS ai_campaign_progress CASCADE;

-- 4. If you need to remove sequence progress columns:
-- ALTER TABLE sequences DROP COLUMN IF EXISTS total_contacts;
-- ALTER TABLE sequences DROP COLUMN IF EXISTS active_contacts;
-- ALTER TABLE sequences DROP COLUMN IF EXISTS completed_contacts;
-- ALTER TABLE sequences DROP COLUMN IF EXISTS failed_contacts;
-- ALTER TABLE sequences DROP COLUMN IF EXISTS progress_percentage;
-- ALTER TABLE sequences DROP COLUMN IF EXISTS last_activity_at;
-- ALTER TABLE sequences DROP COLUMN IF EXISTS estimated_completion_at;

-- 5. Check for any invalid data that might cause crashes
SELECT 'Checking for null or invalid campaign dates...' as check_type;
SELECT id, title, campaign_date FROM campaigns WHERE campaign_date IS NULL;

SELECT 'Checking for invalid timestamps in messages...' as check_type;
SELECT COUNT(*) as invalid_timestamps 
FROM whatsapp_messages 
WHERE timestamp > (EXTRACT(EPOCH FROM NOW()) + 31536000)::BIGINT
   OR timestamp < 0;

-- 6. Create a backup of critical tables before making changes
-- CREATE TABLE campaigns_backup AS SELECT * FROM campaigns;
-- CREATE TABLE sequences_backup AS SELECT * FROM sequences;
