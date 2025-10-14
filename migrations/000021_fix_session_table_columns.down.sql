-- Revert changes to ai_whatsapp_session_nodepath table

-- Remove indexes
ALTER TABLE ai_whatsapp_session_nodepath
  DROP INDEX IF EXISTS idx_ai_session_locked,
  DROP INDEX IF EXISTS idx_ai_session_activity;

-- Remove added columns
ALTER TABLE ai_whatsapp_session_nodepath 
  DROP COLUMN IF EXISTS locked_at,
  DROP COLUMN IF EXISTS last_activity,
  DROP COLUMN IF EXISTS created_at;

-- Revert column names
ALTER TABLE ai_whatsapp_session_nodepath 
  CHANGE COLUMN phone_number id_prospect VARCHAR(255) NOT NULL,
  CHANGE COLUMN device_id id_device VARCHAR(255) NOT NULL;

-- Restore original unique key
ALTER TABLE ai_whatsapp_session_nodepath
  DROP KEY IF EXISTS uniq_ai_whatsapp_session,
  ADD UNIQUE KEY uniq_ai_whatsapp_session (id_prospect, id_device);

-- wasapBot_session_nodepath table was not modified