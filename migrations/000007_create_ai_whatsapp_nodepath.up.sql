-- Create ai_whatsapp_nodepath table for AI WhatsApp conversations
CREATE TABLE IF NOT EXISTS ai_whatsapp_nodepath (
    id_prospect INT AUTO_INCREMENT PRIMARY KEY,
    id_staff VARCHAR(255) NOT NULL,
    prospect_num VARCHAR(255) NOT NULL UNIQUE,
    stage VARCHAR(255) DEFAULT NULL,
    date_order DATETIME DEFAULT NULL,
    conv_last JSON DEFAULT NULL,
    conv_current TEXT DEFAULT NULL,
    jam VARCHAR(255) DEFAULT NULL,
    intro VARCHAR(255) DEFAULT NULL,
    human INT DEFAULT 0 COMMENT '0 = AI active, 1 = human takeover',
    catatan_staff VARCHAR(255) DEFAULT NULL,
    balas INT DEFAULT 0,
    data_image VARCHAR(255) DEFAULT NULL,
    conv_stage TEXT DEFAULT NULL,
    niche VARCHAR(255) DEFAULT NULL,
    bot_balas DATETIME DEFAULT NULL,
    keywordiklan VARCHAR(255) DEFAULT NULL,
    marketer VARCHAR(255) DEFAULT NULL,
    update_today DATETIME DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_prospect_num (prospect_num),
    INDEX idx_id_staff (id_staff),
    INDEX idx_stage (stage),
    INDEX idx_human (human),
    INDEX idx_niche (niche),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Add foreign key constraints if related tables exist
-- Note: Uncomment these if you have the referenced tables
-- ALTER TABLE ai_whatsapp_nodepath ADD CONSTRAINT fk_ai_whatsapp_staff 
--     FOREIGN KEY (id_staff) REFERENCES staff_table(id) ON DELETE CASCADE;