-- Rollback: Rename id_device column back to id_staff in ai_whatsapp_nodepath table
-- This migration reverts the column name change for rollback purposes

ALTER TABLE ai_whatsapp_nodepath 
CHANGE COLUMN id_device id_staff VARCHAR(255) NOT NULL COMMENT 'Staff ID for AI WhatsApp conversations';

-- Update the index to reflect the original column name
DROP INDEX idx_id_device ON ai_whatsapp_nodepath;
CREATE INDEX idx_id_staff ON ai_whatsapp_nodepath(id_staff);

-- Verify the rollback
SELECT 'Successfully rolled back id_device to id_staff in ai_whatsapp_nodepath table' AS status;