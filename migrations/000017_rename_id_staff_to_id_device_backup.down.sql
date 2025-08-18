-- Rollback: Rename id_device column back to id_staff in conversation_log_nodepath_backup table
-- This migration reverts the column name change for rollback purposes

-- Check if the backup table exists before attempting to modify it
SET @table_exists = (SELECT COUNT(*) FROM information_schema.tables 
                    WHERE table_schema = DATABASE() 
                    AND table_name = 'conversation_log_nodepath_backup');

SET @sql = IF(@table_exists > 0, 
    'ALTER TABLE conversation_log_nodepath_backup CHANGE COLUMN id_device id_staff VARCHAR(255) NOT NULL COMMENT "Staff ID for conversation backup tracking"',
    'SELECT "Table conversation_log_nodepath_backup does not exist, skipping rollback" AS status');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Update the index if table exists
SET @index_sql = IF(@table_exists > 0, 
    'DROP INDEX IF EXISTS idx_id_device ON conversation_log_nodepath_backup; CREATE INDEX idx_id_staff ON conversation_log_nodepath_backup(id_staff)',
    'SELECT "Skipping index rollback for non-existent table" AS status');

PREPARE stmt FROM @index_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Verify the rollback
SELECT 'Rollback completed for conversation_log_nodepath_backup table' AS status;