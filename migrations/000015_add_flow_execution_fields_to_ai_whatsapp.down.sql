-- Remove flow execution fields from ai_whatsapp_nodepath table
-- This reverts the consolidation changes

-- Drop the new indexes
DROP INDEX IF EXISTS idx_ai_whatsapp_flow_reference ON ai_whatsapp_nodepath;
DROP INDEX IF EXISTS idx_ai_whatsapp_current_node ON ai_whatsapp_nodepath;
DROP INDEX IF EXISTS idx_ai_whatsapp_execution_status ON ai_whatsapp_nodepath;
DROP INDEX IF EXISTS idx_ai_whatsapp_execution_id ON ai_whatsapp_nodepath;

-- Remove the flow execution columns
ALTER TABLE ai_whatsapp_nodepath 
DROP COLUMN IF EXISTS flow_reference,
DROP COLUMN IF EXISTS current_node,
DROP COLUMN IF EXISTS variables,
DROP COLUMN IF EXISTS execution_status,
DROP COLUMN IF EXISTS execution_id;

-- Revert id_device back to id_staff
ALTER TABLE ai_whatsapp_nodepath CHANGE COLUMN id_device id_staff VARCHAR(255) NOT NULL;

-- Restore the original index
DROP INDEX IF EXISTS idx_id_device ON ai_whatsapp_nodepath;
CREATE INDEX idx_id_staff ON ai_whatsapp_nodepath(id_staff);