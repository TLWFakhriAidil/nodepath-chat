-- Drop remaining old column from chatbot_flows_nodepath table
ALTER TABLE chatbot_flows_nodepath DROP COLUMN global_instance;

-- Show the updated table structure
DESCRIBE chatbot_flows_nodepath;