-- Update chatbot_flows_nodepath table schema
-- Remove old columns and add new ones for device selection and niche

-- Add new columns
ALTER TABLE chatbot_flows_nodepath 
ADD COLUMN IF NOT EXISTS selected_device_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS niche TEXT;

-- Remove old columns
ALTER TABLE chatbot_flows_nodepath 
DROP COLUMN IF EXISTS instance,
DROP COLUMN IF EXISTS open_router_key;

-- Update any existing records to have empty values for new columns
-- (This is safe since we're transitioning from old to new schema)
UPDATE chatbot_flows_nodepath 
SET selected_device_id = '', niche = '' 
WHERE selected_device_id IS NULL OR niche IS NULL;