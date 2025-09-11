-- ============================================
-- AI WhatsApp Table Column Migration Script
-- ============================================
-- This script will:
-- 1. Drop unwanted columns from ai_whatsapp_nodepath
-- 2. Add missing columns based on the Go model
-- 3. Ensure proper data types for all columns
-- ============================================

USE admin_railway;

-- Show current table structure
SELECT '📊 Current Table Structure:' AS Status;
DESCRIBE ai_whatsapp_nodepath;

-- ============================================
-- STEP 1: DROP UNWANTED COLUMNS
-- ============================================
SELECT '🗑️ Step 1: Dropping unwanted columns...' AS Status;

-- Drop jam column if exists
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'jam'
    ),
    'ALTER TABLE ai_whatsapp_nodepath DROP COLUMN jam',
    'SELECT "Column jam does not exist" AS Info'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Drop catatan_staff column if exists
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'catatan_staff'
    ),
    'ALTER TABLE ai_whatsapp_nodepath DROP COLUMN catatan_staff',
    'SELECT "Column catatan_staff does not exist" AS Info'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Drop data_image column if exists
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'data_image'
    ),
    'ALTER TABLE ai_whatsapp_nodepath DROP COLUMN data_image',
    'SELECT "Column data_image does not exist" AS Info'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Drop conv_stage column if exists
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'conv_stage'
    ),
    'ALTER TABLE ai_whatsapp_nodepath DROP COLUMN conv_stage',
    'SELECT "Column conv_stage does not exist" AS Info'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Drop variables column if exists
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'variables'
    ),
    'ALTER TABLE ai_whatsapp_nodepath DROP COLUMN variables',
    'SELECT "Column variables does not exist" AS Info'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Drop bot_balas column if exists
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'bot_balas'
    ),
    'ALTER TABLE ai_whatsapp_nodepath DROP COLUMN bot_balas',
    'SELECT "Column bot_balas does not exist" AS Info'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Drop current_node column if exists
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'current_node'
    ),
    'ALTER TABLE ai_whatsapp_nodepath DROP COLUMN current_node',
    'SELECT "Column current_node does not exist" AS Info'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ============================================
-- STEP 2: ADD/MODIFY REQUIRED COLUMNS
-- ============================================
SELECT '➕ Step 2: Adding/Modifying required columns...' AS Status;

-- Ensure id_prospect is INT(11) and PRIMARY KEY
ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN id_prospect INT(11) NOT NULL AUTO_INCREMENT;

-- Add/Modify flow_reference
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'flow_reference'
    ),
    'ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN flow_reference VARCHAR(255) DEFAULT NULL',
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN flow_reference VARCHAR(255) DEFAULT NULL COMMENT "Reference to chatbot flow being executed"'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add/Modify execution_id
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'execution_id'
    ),
    'ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN execution_id VARCHAR(255) DEFAULT NULL',
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN execution_id VARCHAR(255) DEFAULT NULL COMMENT "Unique execution identifier"'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add/Modify date_order
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'date_order'
    ),
    'ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN date_order DATETIME DEFAULT NULL',
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN date_order DATETIME DEFAULT NULL'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add/Modify id_device
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'id_device'
    ),
    'ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN id_device VARCHAR(255) DEFAULT NULL',
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN id_device VARCHAR(255) DEFAULT NULL'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add/Modify niche
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'niche'
    ),
    'ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN niche VARCHAR(255) DEFAULT NULL',
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN niche VARCHAR(255) DEFAULT NULL'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add/Modify prospect_name
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'prospect_name'
    ),
    'ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN prospect_name VARCHAR(255) DEFAULT NULL',
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN prospect_name VARCHAR(255) DEFAULT NULL'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add/Modify prospect_num
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'prospect_num'
    ),
    'ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN prospect_num VARCHAR(255) DEFAULT NULL',
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN prospect_num VARCHAR(255) DEFAULT NULL'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add/Modify intro
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'intro'
    ),
    'ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN intro VARCHAR(255) DEFAULT NULL',
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN intro VARCHAR(255) DEFAULT NULL'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add/Modify stage
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'stage'
    ),
    'ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN stage VARCHAR(255) DEFAULT NULL',
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN stage VARCHAR(255) DEFAULT NULL'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add/Modify conv_last
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'conv_last'
    ),
    'ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN conv_last TEXT DEFAULT NULL',
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN conv_last TEXT DEFAULT NULL'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add/Modify conv_current
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'conv_current'
    ),
    'ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN conv_current TEXT DEFAULT NULL',
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN conv_current TEXT DEFAULT NULL'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add/Modify execution_status
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'execution_status'
    ),
    'SELECT "Column execution_status already exists" AS Info',
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN execution_status ENUM("active","completed","failed") DEFAULT NULL COMMENT "Flow execution status"'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add/Modify flow_id
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'flow_id'
    ),
    'ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN flow_id VARCHAR(255) DEFAULT NULL',
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN flow_id VARCHAR(255) DEFAULT NULL COMMENT "ID of the current chatbot flow"'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add/Modify current_node_id
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'current_node_id'
    ),
    'ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN current_node_id VARCHAR(255) DEFAULT NULL',
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN current_node_id VARCHAR(255) DEFAULT NULL COMMENT "Current node ID in the chatbot flow"'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add/Modify last_node_id
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'last_node_id'
    ),
    'ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN last_node_id VARCHAR(255) DEFAULT NULL',
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN last_node_id VARCHAR(255) DEFAULT NULL COMMENT "Previous node ID for flow tracking"'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add/Modify waiting_for_reply
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'waiting_for_reply'
    ),
    'ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN waiting_for_reply TINYINT(1) DEFAULT 0',
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN waiting_for_reply TINYINT(1) DEFAULT 0 COMMENT "1 = waiting for user reply"'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add/Modify balas
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'balas'
    ),
    'ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN balas VARCHAR(255) DEFAULT NULL',
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN balas VARCHAR(255) DEFAULT NULL'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add/Modify human
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'human'
    ),
    'ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN human INT(11) DEFAULT 0',
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN human INT(11) DEFAULT 0 COMMENT "0 = AI active, 1 = human takeover"'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add/Modify keywordiklan
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'keywordiklan'
    ),
    'ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN keywordiklan VARCHAR(255) DEFAULT NULL',
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN keywordiklan VARCHAR(255) DEFAULT NULL'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add/Modify marketer
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'marketer'
    ),
    'ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN marketer VARCHAR(255) DEFAULT NULL',
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN marketer VARCHAR(255) DEFAULT NULL'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add/Modify created_at
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'created_at'
    ),
    'ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP',
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add/Modify updated_at
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'updated_at'
    ),
    'ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP',
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add/Modify update_today
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND COLUMN_NAME = 'update_today'
    ),
    'ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN update_today DATETIME DEFAULT NULL',
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN update_today DATETIME DEFAULT NULL'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ============================================
-- STEP 3: ADD INDEXES
-- ============================================
SELECT '🔍 Step 3: Adding indexes...' AS Status;

-- Add unique index on execution_id if not exists
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.STATISTICS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND INDEX_NAME = 'uniq_execution_id'
    ),
    'SELECT "Index uniq_execution_id already exists" AS Info',
    'CREATE UNIQUE INDEX uniq_execution_id ON ai_whatsapp_nodepath(execution_id)'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add index on flow_id if not exists
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.STATISTICS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND INDEX_NAME = 'idx_flow_id'
    ),
    'SELECT "Index idx_flow_id already exists" AS Info',
    'CREATE INDEX idx_flow_id ON ai_whatsapp_nodepath(flow_id)'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add index on current_node_id if not exists
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.STATISTICS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND INDEX_NAME = 'idx_current_node_id'
    ),
    'SELECT "Index idx_current_node_id already exists" AS Info',
    'CREATE INDEX idx_current_node_id ON ai_whatsapp_nodepath(current_node_id)'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add index on id_device if not exists
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.STATISTICS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND INDEX_NAME = 'idx_id_device'
    ),
    'SELECT "Index idx_id_device already exists" AS Info',
    'CREATE INDEX idx_id_device ON ai_whatsapp_nodepath(id_device)'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add index on prospect_num if not exists
SET @sql = (SELECT IF(
    EXISTS(
        SELECT * FROM INFORMATION_SCHEMA.STATISTICS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath' 
        AND INDEX_NAME = 'idx_prospect_num'
    ),
    'SELECT "Index idx_prospect_num already exists" AS Info',
    'CREATE INDEX idx_prospect_num ON ai_whatsapp_nodepath(prospect_num)'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ============================================
-- FINAL VERIFICATION
-- ============================================
SELECT '✅ Migration completed! Final table structure:' AS Status;
DESCRIBE ai_whatsapp_nodepath;

-- Show column count
SELECT CONCAT('Total columns: ', COUNT(*)) AS Summary
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_SCHEMA = DATABASE() 
AND TABLE_NAME = 'ai_whatsapp_nodepath';

-- List all columns
SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT, COLUMN_COMMENT
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_SCHEMA = DATABASE() 
AND TABLE_NAME = 'ai_whatsapp_nodepath'
ORDER BY ORDINAL_POSITION;
