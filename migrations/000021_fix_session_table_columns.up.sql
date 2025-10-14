-- Add missing columns to ai_whatsapp_session_nodepath table
-- The repository code expects phone_number and device_id columns with locking capabilities

-- First, check if we need to rename columns
ALTER TABLE ai_whatsapp_session_nodepath 
  CHANGE COLUMN id_prospect phone_number VARCHAR(255) NOT NULL,
  CHANGE COLUMN id_device device_id VARCHAR(255) NOT NULL;

-- Add the missing columns for session locking
ALTER TABLE ai_whatsapp_session_nodepath 
  ADD COLUMN IF NOT EXISTS locked_at TIMESTAMP NULL DEFAULT NULL,
  ADD COLUMN IF NOT EXISTS last_activity TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  ADD COLUMN IF NOT EXISTS created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- Add indexes for better performance
ALTER TABLE ai_whatsapp_session_nodepath
  ADD INDEX IF NOT EXISTS idx_ai_session_locked (locked_at),
  ADD INDEX IF NOT EXISTS idx_ai_session_activity (last_activity);

-- Update the unique key to use new column names
ALTER TABLE ai_whatsapp_session_nodepath
  DROP KEY IF EXISTS uniq_ai_whatsapp_session,
  ADD UNIQUE KEY uniq_ai_whatsapp_session (phone_number, device_id);

-- wasapBot_session_nodepath table keeps its original columns 
-- since it's used differently by wasapbot_repository.go with id_prospect and id_device