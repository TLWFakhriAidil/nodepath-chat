-- Add flow tracking fields to ai_whatsapp_nodepath table
-- These fields enable proper user reply node handling and flow continuation

-- First, add the new columns
ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN current_node_id VARCHAR(255) DEFAULT NULL COMMENT 'Current node ID in the chatbot flow';

ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN waiting_for_reply TINYINT(1) DEFAULT 0 COMMENT '1 = waiting for user reply, 0 = not waiting';

ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN flow_id VARCHAR(255) DEFAULT NULL COMMENT 'ID of the current chatbot flow being executed';

ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN last_node_id VARCHAR(255) DEFAULT NULL COMMENT 'Previous node ID for flow tracking';

-- Then add indexes for better performance
CREATE INDEX idx_current_node_id ON ai_whatsapp_nodepath(current_node_id);
CREATE INDEX idx_waiting_for_reply ON ai_whatsapp_nodepath(waiting_for_reply);
CREATE INDEX idx_flow_id ON ai_whatsapp_nodepath(flow_id);

-- Verify the new columns were added
SELECT 'Successfully added flow tracking fields to ai_whatsapp_nodepath table' AS status;