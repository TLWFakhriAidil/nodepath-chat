-- Production Migration: Remove deprecated columns from ai_whatsapp_nodepath table
-- This script removes deprecated columns that are no longer needed

-- Remove deprecated columns one by one with existence checks

-- 1. Drop jam column (deprecated)
SET @col_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND COLUMN_NAME = 'jam'
);
SET @sql = IF(@col_exists > 0, 
    'ALTER TABLE ai_whatsapp_nodepath DROP COLUMN jam',
    'SELECT "Column jam already removed" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 2. Add intro column
SET @col_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND COLUMN_NAME = 'intro'
);
SET @sql = IF(@col_exists = 0, 
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN intro VARCHAR(255) DEFAULT NULL',
    'SELECT "Column intro already exists" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 3. Add date_order column
SET @col_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND COLUMN_NAME = 'date_order'
);
SET @sql = IF(@col_exists = 0, 
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN date_order TIMESTAMP NULL',
    'SELECT "Column date_order already exists" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 4. Add balas column (rename bot_balas if needed)
SET @col_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND COLUMN_NAME = 'balas'
);
SET @sql = IF(@col_exists = 0, 
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN balas TINYINT(1) DEFAULT 1',
    'SELECT "Column balas already exists" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 5. Drop data_image column (deprecated)
SET @col_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND COLUMN_NAME = 'data_image'
);
SET @sql = IF(@col_exists > 0, 
    'ALTER TABLE ai_whatsapp_nodepath DROP COLUMN data_image',
    'SELECT "Column data_image already removed" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 6. Drop conv_stage column (deprecated)
SET @col_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND COLUMN_NAME = 'conv_stage'
);
SET @sql = IF(@col_exists > 0, 
    'ALTER TABLE ai_whatsapp_nodepath DROP COLUMN conv_stage',
    'SELECT "Column conv_stage already removed" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 7. Add keywordiklan column
SET @col_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND COLUMN_NAME = 'keywordiklan'
);
SET @sql = IF(@col_exists = 0, 
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN keywordiklan VARCHAR(255)',
    'SELECT "Column keywordiklan already exists" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 8. Add marketer column
SET @col_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND COLUMN_NAME = 'marketer'
);
SET @sql = IF(@col_exists = 0, 
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN marketer VARCHAR(255)',
    'SELECT "Column marketer already exists" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 9. Add update_today column
SET @col_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND COLUMN_NAME = 'update_today'
);
SET @sql = IF(@col_exists = 0, 
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN update_today TIMESTAMP NULL',
    'SELECT "Column update_today already exists" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 10. Drop bot_balas column (deprecated)
SET @col_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND COLUMN_NAME = 'bot_balas'
);
SET @sql = IF(@col_exists > 0, 
    'ALTER TABLE ai_whatsapp_nodepath DROP COLUMN bot_balas',
    'SELECT "Column bot_balas already removed" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 11. Drop variables column (deprecated)
SET @col_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND COLUMN_NAME = 'variables'
);
SET @sql = IF(@col_exists > 0, 
    'ALTER TABLE ai_whatsapp_nodepath DROP COLUMN variables',
    'SELECT "Column variables already removed" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 12. Drop catatan_staff column (deprecated)
SET @col_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND COLUMN_NAME = 'catatan_staff'
);
SET @sql = IF(@col_exists > 0, 
    'ALTER TABLE ai_whatsapp_nodepath DROP COLUMN catatan_staff',
    'SELECT "Column catatan_staff already removed" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 13. Drop current_node column (deprecated)
SET @col_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND COLUMN_NAME = 'current_node'
);
SET @sql = IF(@col_exists > 0, 
    'ALTER TABLE ai_whatsapp_nodepath DROP COLUMN current_node',
    'SELECT "Column current_node already removed" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 14. Add new schema columns
-- Add flow_reference column
SET @col_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND COLUMN_NAME = 'flow_reference'
);
SET @sql = IF(@col_exists = 0, 
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN flow_reference VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT "Reference to chatbot flow being executed"',
    'SELECT "Column flow_reference already exists" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Add execution_id column
SET @col_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND COLUMN_NAME = 'execution_id'
);
SET @sql = IF(@col_exists = 0, 
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN execution_id VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT "Unique execution identifier"',
    'SELECT "Column execution_id already exists" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Add prospect_name column
SET @col_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND COLUMN_NAME = 'prospect_name'
);
SET @sql = IF(@col_exists = 0, 
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN prospect_name VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL',
    'SELECT "Column prospect_name already exists" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Add execution_status column
SET @col_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND COLUMN_NAME = 'execution_status'
);
SET @sql = IF(@col_exists = 0, 
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN execution_status ENUM("active","completed","failed") COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT "Flow execution status"',
    'SELECT "Column execution_status already exists" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Add flow_id column
SET @col_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND COLUMN_NAME = 'flow_id'
);
SET @sql = IF(@col_exists = 0, 
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN flow_id VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT "ID of the current chatbot flow being executed"',
    'SELECT "Column flow_id already exists" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Add current_node_id column
SET @col_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND COLUMN_NAME = 'current_node_id'
);
SET @sql = IF(@col_exists = 0, 
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN current_node_id VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT "Current node ID in the chatbot flow"',
    'SELECT "Column current_node_id already exists" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Add last_node_id column
SET @col_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND COLUMN_NAME = 'last_node_id'
);
SET @sql = IF(@col_exists = 0, 
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN last_node_id VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT "Previous node ID for flow tracking"',
    'SELECT "Column last_node_id already exists" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Add waiting_for_reply column
SET @col_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND COLUMN_NAME = 'waiting_for_reply'
);
SET @sql = IF(@col_exists = 0, 
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN waiting_for_reply TINYINT(1) DEFAULT 0 COMMENT "1 = waiting for user reply, 0 = not waiting"',
    'SELECT "Column waiting_for_reply already exists" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 15. Modify id_prospect to INT if it exists as VARCHAR
SET @col_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND COLUMN_NAME = 'id_prospect'
    AND DATA_TYPE = 'varchar'
);
SET @sql = IF(@col_exists > 0, 
    'ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN id_prospect INT NOT NULL',
    'SELECT "Column id_prospect already correct type" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Verify all columns exist
SELECT 
    COLUMN_NAME,
    DATA_TYPE,
    IS_NULLABLE,
    COLUMN_DEFAULT
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_SCHEMA = DATABASE() 
AND TABLE_NAME = 'ai_whatsapp_nodepath' 
ORDER BY ORDINAL_POSITION;

-- Create indexes for new columns
-- Index for execution_id (unique)
SET @index_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.STATISTICS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND INDEX_NAME = 'uniq_execution_id'
);
SET @sql = IF(@index_exists = 0, 
    'ALTER TABLE ai_whatsapp_nodepath ADD UNIQUE KEY uniq_execution_id (execution_id)',
    'SELECT "Index uniq_execution_id already exists" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Index for flow_id
SET @index_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.STATISTICS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND INDEX_NAME = 'idx_flow_id'
);
SET @sql = IF(@index_exists = 0, 
    'ALTER TABLE ai_whatsapp_nodepath ADD KEY idx_flow_id (flow_id)',
    'SELECT "Index idx_flow_id already exists" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Index for current_node_id
SET @index_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.STATISTICS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND INDEX_NAME = 'idx_current_node_id'
);
SET @sql = IF(@index_exists = 0, 
    'ALTER TABLE ai_whatsapp_nodepath ADD KEY idx_current_node_id (current_node_id)',
    'SELECT "Index idx_current_node_id already exists" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Index for id_device
SET @index_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.STATISTICS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND INDEX_NAME = 'idx_id_device'
);
SET @sql = IF(@index_exists = 0, 
    'ALTER TABLE ai_whatsapp_nodepath ADD KEY idx_id_device (id_device)',
    'SELECT "Index idx_id_device already exists" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Index for prospect_num
SET @index_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.STATISTICS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND INDEX_NAME = 'idx_prospect_num'
);
SET @sql = IF(@index_exists = 0, 
    'ALTER TABLE ai_whatsapp_nodepath ADD KEY idx_prospect_num (prospect_num)',
    'SELECT "Index idx_prospect_num already exists" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Show success message
SELECT 'Production database schema migration completed successfully - deprecated columns removed and new schema applied' AS status;