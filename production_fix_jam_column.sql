-- Production Migration: Add missing 'jam' column to ai_whatsapp_nodepath table
-- This script fixes the "Unknown column 'jam' in 'field list'" error in Railway production

-- Check if the column exists before adding it
SET @column_exists = (
    SELECT COUNT(*)
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'ai_whatsapp_nodepath'
    AND COLUMN_NAME = 'jam'
);

-- Add the jam column if it doesn't exist
SET @sql = IF(@column_exists = 0,
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN jam VARCHAR(255) DEFAULT NULL',
    'SELECT "Column jam already exists" AS message'
);

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Verify the column was added
SELECT 
    COLUMN_NAME,
    DATA_TYPE,
    IS_NULLABLE,
    COLUMN_DEFAULT
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
AND TABLE_NAME = 'ai_whatsapp_nodepath'
AND COLUMN_NAME = 'jam';

-- Show success message
SELECT 'Production database schema fix completed successfully' AS status;