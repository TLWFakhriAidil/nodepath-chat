-- Production Migration: Add missing columns to ai_whatsapp_nodepath table
-- This script fixes the "Unknown column" errors in Railway production

-- Add missing columns one by one with existence checks

-- 1. Add jam column
SET @col_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND COLUMN_NAME = 'jam'
);
SET @sql = IF(@col_exists = 0, 
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN jam VARCHAR(255) DEFAULT NULL',
    'SELECT "Column jam already exists" AS message'
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

-- 5. Add data_image column
SET @col_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND COLUMN_NAME = 'data_image'
);
SET @sql = IF(@col_exists = 0, 
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN data_image TEXT',
    'SELECT "Column data_image already exists" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 6. Add conv_stage column
SET @col_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND COLUMN_NAME = 'conv_stage'
);
SET @sql = IF(@col_exists = 0, 
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN conv_stage VARCHAR(255)',
    'SELECT "Column conv_stage already exists" AS message'
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

-- 10. Modify bot_balas to TIMESTAMP if it exists as TINYINT
SET @col_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'ai_whatsapp_nodepath' 
    AND COLUMN_NAME = 'bot_balas'
    AND DATA_TYPE = 'tinyint'
);
SET @sql = IF(@col_exists > 0, 
    'ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN bot_balas TIMESTAMP NULL',
    'SELECT "Column bot_balas already correct type" AS message'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 11. Modify id_prospect to INT if it exists as VARCHAR
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

-- Show success message
SELECT 'Production database schema fix completed successfully' AS status;