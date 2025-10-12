-- Make device_id column nullable to support manual device creation
-- This fixes the issue where WAHA provider requires device_id to be empty initially

-- First check if the column is NOT NULL, then alter it
ALTER TABLE device_setting_nodepath MODIFY COLUMN device_id VARCHAR(255) NULL;