-- Rollback migration - Re-add dropped columns if needed
-- Note: This will restore the columns but not the data

ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS jam VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
ADD COLUMN IF NOT EXISTS catatan_staff VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
ADD COLUMN IF NOT EXISTS data_image VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
ADD COLUMN IF NOT EXISTS conv_stage TEXT COLLATE utf8mb4_unicode_ci,
ADD COLUMN IF NOT EXISTS variables TEXT COLLATE utf8mb4_unicode_ci COMMENT 'Flow execution variables',
ADD COLUMN IF NOT EXISTS bot_balas TIMESTAMP NULL DEFAULT NULL COMMENT 'Bot reply timestamp',
ADD COLUMN IF NOT EXISTS current_node VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Current node in the flow execution';
