-- Rollback: Make device_id column NOT NULL again
ALTER TABLE device_setting_nodepath MODIFY COLUMN device_id VARCHAR(255) NOT NULL;