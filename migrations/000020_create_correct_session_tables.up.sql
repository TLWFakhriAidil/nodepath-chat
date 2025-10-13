-- Create session management tables with correct names
-- These tables prevent duplicate message processing by maintaining session locks

CREATE TABLE IF NOT EXISTS ai_whatsapp_session (
    phone_number VARCHAR(255) NOT NULL,
    device_id VARCHAR(255) NOT NULL,
    locked_at TIMESTAMP NULL DEFAULT NULL,
    last_activity TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (phone_number, device_id),
    INDEX idx_ai_session_locked (locked_at),
    INDEX idx_ai_session_activity (last_activity),
    INDEX idx_ai_session_device (device_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS wasapbot_session (
    phone_number VARCHAR(255) NOT NULL,
    device_id VARCHAR(255) NOT NULL,
    locked_at TIMESTAMP NULL DEFAULT NULL,
    last_activity TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (phone_number, device_id),
    INDEX idx_wasap_session_locked (locked_at),
    INDEX idx_wasap_session_activity (last_activity),
    INDEX idx_wasap_session_device (device_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;