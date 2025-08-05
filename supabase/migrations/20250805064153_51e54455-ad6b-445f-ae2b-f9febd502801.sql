-- Add AI prompt columns to chatbot_flows_nodepath table
ALTER TABLE chatbot_flows_nodepath 
ADD COLUMN IF NOT EXISTS system_prompt TEXT,
ADD COLUMN IF NOT EXISTS instance VARCHAR(255),
ADD COLUMN IF NOT EXISTS open_router_key VARCHAR(255);