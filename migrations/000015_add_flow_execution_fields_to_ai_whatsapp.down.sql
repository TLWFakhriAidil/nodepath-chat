-- Remove new schema fields from ai_whatsapp_nodepath table
-- This reverts the schema update changes
-- Note: Deprecated columns (current_node, variables) are permanently removed and will not be restored

-- Drop the new indexes
DROP INDEX IF EXISTS idx_ai_whatsapp_flow_reference ON ai_whatsapp_nodepath;
DROP INDEX IF EXISTS idx_ai_whatsapp_execution_status ON ai_whatsapp_nodepath;
DROP INDEX IF EXISTS idx_ai_whatsapp_execution_id ON ai_whatsapp_nodepath;
DROP INDEX IF EXISTS idx_flow_id ON ai_whatsapp_nodepath;
DROP INDEX IF EXISTS idx_current_node_id ON ai_whatsapp_nodepath;
DROP INDEX IF EXISTS idx_prospect_num ON ai_whatsapp_nodepath;

-- Remove the new schema columns
ALTER TABLE ai_whatsapp_nodepath 
DROP COLUMN IF EXISTS flow_reference,
DROP COLUMN IF EXISTS execution_status,
DROP COLUMN IF EXISTS execution_id,
DROP COLUMN IF EXISTS flow_id,
DROP COLUMN IF EXISTS current_node_id,
DROP COLUMN IF EXISTS last_node_id,
DROP COLUMN IF EXISTS waiting_for_reply,
DROP COLUMN IF EXISTS prospect_name;

-- Note: Deprecated columns (current_node, variables) are not restored as they are permanently removed
SELECT 'Migration rollback completed. Deprecated columns (current_node, variables) remain permanently removed.' AS status;

-- Revert id_device back to id_staff
ALTER TABLE ai_whatsapp_nodepath CHANGE COLUMN id_device id_staff VARCHAR(255) NOT NULL;

-- Restore the original index
DROP INDEX IF EXISTS idx_id_device ON ai_whatsapp_nodepath;
CREATE INDEX idx_id_staff ON ai_whatsapp_nodepath(id_staff);