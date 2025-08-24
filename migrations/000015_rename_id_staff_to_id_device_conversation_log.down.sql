-- Rollback: Rename id_device column back to id_staff in conversation_log_nodepath table
-- This migration reverts the column name change for rollback purposes

ALTER TABLE conversation_log_nodepath 
CHANGE COLUMN id_device id_staff VARCHAR(255) NOT NULL COMMENT 'Staff ID for conversation tracking';

-- Update the index to reflect the original column name
DROP INDEX idx_id_device ON conversation_log_nodepath;
CREATE INDEX idx_id_staff ON conversation_log_nodepath(id_staff);

-- Verify the rollback
SELECT 'Successfully rolled back id_device to id_staff in conversation_log_nodepath table' AS status;