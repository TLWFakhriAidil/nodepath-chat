-- Rename id_staff column to id_device in ai_whatsapp_nodepath table
-- This migration ensures consistency with device-based identification across the system

ALTER TABLE ai_whatsapp_nodepath 
CHANGE COLUMN id_staff id_device VARCHAR(255) NOT NULL COMMENT 'Device ID for AI WhatsApp conversations';

-- Update the index to reflect the new column name
DROP INDEX idx_id_staff ON ai_whatsapp_nodepath;
CREATE INDEX idx_id_device ON ai_whatsapp_nodepath(id_device);

-- Verify the change
SELECT 'Successfully renamed id_staff to id_device in ai_whatsapp_nodepath table' AS status;