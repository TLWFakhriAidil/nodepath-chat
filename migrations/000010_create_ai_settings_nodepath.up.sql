-- Create ai_settings_nodepath table for AI prompt configurations
CREATE TABLE IF NOT EXISTS ai_settings_nodepath (
    id VARCHAR(255) PRIMARY KEY,
    id_staff VARCHAR(255) NOT NULL,
    system_prompt TEXT NOT NULL,
    closing_prompt TEXT DEFAULT NULL,
    instance_prompt TEXT DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_id_staff (id_staff),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default AI settings
INSERT IGNORE INTO ai_settings_nodepath (id, id_staff, system_prompt, closing_prompt, instance_prompt) VALUES
('default-ai-settings', 'default', 
'You are a helpful AI assistant for WhatsApp conversations. Follow the conversation flow and respond appropriately to user messages.',
'Thank you for your time. If you have any other questions, feel free to ask!',
'Please respond in a friendly and professional manner.');