-- Fix missing 'jam' column in ai_whatsapp_nodepath table
-- This script will add the 'jam' column if it doesn't exist

-- Check if the column exists and add it if missing
SET @col_exists = 0;
SELECT COUNT(*) INTO @col_exists 
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_SCHEMA = DATABASE() 
  AND TABLE_NAME = 'ai_whatsapp_nodepath' 
  AND COLUMN_NAME = 'jam';

-- Add the column if it doesn't exist
SET @sql = IF(@col_exists = 0, 
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN jam VARCHAR(255) DEFAULT NULL AFTER conv_current',
    'SELECT "Column jam already exists" as message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Verify the column was added
SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_SCHEMA = DATABASE() 
  AND TABLE_NAME = 'ai_whatsapp_nodepath' 
  AND COLUMN_NAME = 'jam';