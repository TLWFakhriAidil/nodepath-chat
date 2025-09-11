-- Simple MySQL script to DROP unused columns
-- Run each command one by one in your MySQL database

-- 1. First check current table structure
DESCRIBE ai_whatsapp_nodepath;

-- 2. Drop each column individually (run one by one)
ALTER TABLE ai_whatsapp_nodepath DROP COLUMN jam;
ALTER TABLE ai_whatsapp_nodepath DROP COLUMN catatan_staff;
ALTER TABLE ai_whatsapp_nodepath DROP COLUMN data_image;
ALTER TABLE ai_whatsapp_nodepath DROP COLUMN conv_stage;
ALTER TABLE ai_whatsapp_nodepath DROP COLUMN variables;
ALTER TABLE ai_whatsapp_nodepath DROP COLUMN bot_balas;
ALTER TABLE ai_whatsapp_nodepath DROP COLUMN current_node;

-- 3. Verify columns were dropped
SHOW COLUMNS FROM ai_whatsapp_nodepath;
