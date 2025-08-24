-- Rename id_staff column to id_device in conversation_log_nodepath table
-- This migration ensures consistency with device-based identification across the system

ALTER TABLE conversation_log_nodepath 
CHANGE COLUMN id_staff id_device VARCHAR(255) NOT NULL COMMENT 'Device ID for conversation tracking';

-- Update the index to reflect the new column name
DROP INDEX idx_id_staff ON conversation_log_nodepath;
CREATE INDEX idx_id_device ON conversation_log_nodepath(id_device);

-- Verify the change
SELECT 'Successfully renamed id_staff to id_device in conversation_log_nodepath table' AS status;