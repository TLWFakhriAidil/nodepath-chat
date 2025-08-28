-- Update provider from 'rvsb_wasap' to 'waha' in device_setting_nodepath table
-- This migration changes RVSB WASAP provider to WAHA provider

-- Update existing records that have 'rvsb_wasap' provider to 'waha'
UPDATE device_setting_nodepath 
SET provider = 'waha' 
WHERE provider = 'rvsb_wasap';

-- Update the comment to reflect the new provider options
ALTER TABLE device_setting_nodepath 
MODIFY COLUMN provider VARCHAR(255) NOT NULL COMMENT 'whacenter, wablas, waha';