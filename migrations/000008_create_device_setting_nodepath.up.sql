-- Create device_setting_nodepath table for device configurations
CREATE TABLE IF NOT EXISTS device_setting_nodepath (
    id VARCHAR(255) PRIMARY KEY,
    device_id VARCHAR(255) NOT NULL,
    api_key_option VARCHAR(255) NOT NULL COMMENT 'chat_gpt_4o, chat_gpt_5_mini, gemini_pro_15, etc.',
    webhook_id VARCHAR(255) DEFAULT NULL,
    provider VARCHAR(255) NOT NULL COMMENT 'whacenter, wablas, rvsb_wasap',
    phone_number VARCHAR(255) DEFAULT NULL,
    api_key TEXT NOT NULL,
    id_device VARCHAR(255) NOT NULL,
    id_erp VARCHAR(255) DEFAULT NULL,
    id_admin VARCHAR(255) DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY unique_device_id (device_id),
    INDEX idx_device_id (device_id),
    INDEX idx_id_device (id_device),
    INDEX idx_provider (provider),
    INDEX idx_api_key_option (api_key_option),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default settings for special devices SCHQ-S94 and SCHQ-S12
INSERT IGNORE INTO device_setting_nodepath (id, device_id, api_key_option, provider, api_key, id_device) VALUES
('default-schq-s94', 'SCHQ-S94', 'gpt-4.1', 'openai', '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', 'SCHQ-S94'),
('default-schq-s12', 'SCHQ-S12', 'gpt-4.1', 'openai', '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', 'SCHQ-S12');