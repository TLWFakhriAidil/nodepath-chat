-- Create conversation_log_nodepath table for AI conversation history
CREATE TABLE IF NOT EXISTS conversation_log_nodepath (
    id INT AUTO_INCREMENT PRIMARY KEY,
    prospect_num VARCHAR(255) NOT NULL,
    id_staff VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    sender VARCHAR(10) NOT NULL COMMENT 'user or bot',
    stage VARCHAR(255) DEFAULT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_prospect_num (prospect_num),
    INDEX idx_id_staff (id_staff),
    INDEX idx_sender (sender),
    INDEX idx_stage (stage),
    INDEX idx_timestamp (timestamp),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Add foreign key constraint to ai_whatsapp_nodepath if needed
-- ALTER TABLE conversation_log_nodepath ADD CONSTRAINT fk_conversation_log_prospect 
--     FOREIGN KEY (prospect_num) REFERENCES ai_whatsapp_nodepath(prospect_num) ON DELETE CASCADE;