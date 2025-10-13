CREATE TABLE IF NOT EXISTS ai_whatsapp_session_nodepath (
    id_sessionX INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    id_prospect VARCHAR(255) NOT NULL,
    id_device VARCHAR(255) NOT NULL,
    `timestamp` VARCHAR(255) NOT NULL,
    UNIQUE KEY uniq_ai_whatsapp_session (id_prospect, id_device),
    KEY idx_ai_whatsapp_session_device (id_device)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS wasapBot_session_nodepath (
    id_sessionY INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    id_prospect VARCHAR(255) NOT NULL,
    id_device VARCHAR(255) NOT NULL,
    `timestamp` VARCHAR(255) NOT NULL,
    UNIQUE KEY uniq_wasapbot_session (id_prospect, id_device),
    KEY idx_wasapbot_session_device (id_device)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;