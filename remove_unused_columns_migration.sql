-- =====================================================
-- AI WhatsApp NodePath Table Cleanup Migration
-- Removes unused columns: variables, current_node, execution_status
-- =====================================================

-- Step 1: Create backup table with full original schema
CREATE TABLE ai_whatsapp_nodepath_backup AS 
SELECT * FROM ai_whatsapp_nodepath;

-- Step 2: Add comment to backup table for reference
ALTER TABLE ai_whatsapp_nodepath_backup 
COMMENT = 'Backup of ai_whatsapp_nodepath before removing unused columns - Created on migration';

-- Step 3: Verify backup was created successfully
-- (This will show row count - should match original table)
SELECT 
    'Original Table' as table_name, 
    COUNT(*) as row_count 
FROM ai_whatsapp_nodepath
UNION ALL
SELECT 
    'Backup Table' as table_name, 
    COUNT(*) as row_count 
FROM ai_whatsapp_nodepath_backup;

-- Step 4: Drop unused columns from main table
-- Note: These columns are never used in INSERT, UPDATE, or SELECT operations

-- Drop variables column (never used)
ALTER TABLE ai_whatsapp_nodepath 
DROP COLUMN variables;

-- Drop current_node column (legacy, replaced by current_node_id)
ALTER TABLE ai_whatsapp_nodepath 
DROP COLUMN current_node;

-- Drop execution_status column (never used)
ALTER TABLE ai_whatsapp_nodepath 
DROP COLUMN execution_status;

-- Step 5: Verify the cleanup was successful
SELECT 
    'Columns Removed Successfully' as status,
    COUNT(*) as remaining_rows
FROM ai_whatsapp_nodepath;

-- Step 6: Show final schema of cleaned table
DESCRIBE ai_whatsapp_nodepath;

-- =====================================================
-- ROLLBACK INSTRUCTIONS (if needed)
-- =====================================================
-- To rollback this migration, run:
-- 
-- DROP TABLE ai_whatsapp_nodepath;
-- RENAME TABLE ai_whatsapp_nodepath_backup TO ai_whatsapp_nodepath;
-- 
-- This will restore the original table with all columns
-- =====================================================

-- =====================================================
-- FINAL SCHEMA AFTER CLEANUP
-- =====================================================
-- The ai_whatsapp_nodepath table will contain only these columns:
-- 
-- 1. id (Primary Key)
-- 2. id_prospect (Foreign Key)
-- 3. id_device (Device Identifier)
-- 4. prospect_num (Phone Number)
-- 5. date_order (Order Date)
-- 6. niche (Business Niche)
-- 7. intro (Introduction Text)
-- 8. conv_last (Last Conversation Value)
-- 9. conv_current (Current Conversation Value)
-- 10. conv_stage (Conversation Stage)
-- 11. stage (Current Stage)
-- 12. bot_balas (Bot Reply Timestamp)
-- 13. balas (Reply Count/Flag)
-- 14. jam (Time/Hour)
-- 15. human (Human Takeover Flag)
-- 16. keywordiklan (Advertisement Keyword)
-- 17. catatan_staff (Staff Notes)
-- 18. data_image (Image Data)
-- 19. marketer (Marketer Identifier)
-- 20. created_at (Creation Timestamp)
-- 21. updated_at (Update Timestamp)
-- 22. update_today (Today's Update Flag)
-- 23. waiting_for_reply (Waiting for User Reply Flag)
-- 24. flow_id (Flow Identifier)
-- 25. last_node_id (Last Node Identifier)
-- 26. current_node_id (Current Node Identifier)
-- 
-- REMOVED COLUMNS:
-- - variables (never used)
-- - current_node (legacy, replaced by current_node_id)
-- - execution_status (never used)
-- =====================================================