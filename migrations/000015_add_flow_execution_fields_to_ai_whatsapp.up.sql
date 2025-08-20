-- Add flow execution fields to ai_whatsapp_nodepath table
-- This allows consolidating all conversation data into ai_whatsapp_nodepath

ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN flow_reference VARCHAR(255) DEFAULT NULL COMMENT 'Reference to chatbot flow being executed',
ADD COLUMN current_node VARCHAR(255) DEFAULT NULL COMMENT 'Current node in the flow execution',
ADD COLUMN variables JSON DEFAULT NULL COMMENT 'Flow execution variables',
ADD COLUMN execution_status ENUM('active', 'completed', 'failed') DEFAULT NULL COMMENT 'Flow execution status',
ADD COLUMN execution_id VARCHAR(255) DEFAULT NULL COMMENT 'Unique execution identifier';

-- Add indexes for better performance
CREATE INDEX idx_ai_whatsapp_flow_reference ON ai_whatsapp_nodepath(flow_reference);
CREATE INDEX idx_ai_whatsapp_current_node ON ai_whatsapp_nodepath(current_node);
CREATE INDEX idx_ai_whatsapp_execution_status ON ai_whatsapp_nodepath(execution_status);
CREATE INDEX idx_ai_whatsapp_execution_id ON ai_whatsapp_nodepath(execution_id);

-- Update id_staff to id_device for consistency (rename column)
ALTER TABLE ai_whatsapp_nodepath CHANGE COLUMN id_staff id_device VARCHAR(255) NOT NULL;

-- Update the index to use the new column name
DROP INDEX idx_id_staff ON ai_whatsapp_nodepath;
CREATE INDEX idx_id_device ON ai_whatsapp_nodepath(id_device);