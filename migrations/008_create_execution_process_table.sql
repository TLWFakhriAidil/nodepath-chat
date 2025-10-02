-- Create execution_process_nodepath table for preventing duplicate parallel processing
CREATE TABLE IF NOT EXISTS execution_process_nodepath (
    id_chatInput INT AUTO_INCREMENT PRIMARY KEY,
    id_device VARCHAR(255) NOT NULL,
    id_prospect VARCHAR(255) NOT NULL,
    times DATETIME NOT NULL,
    INDEX idx_device_prospect (id_device, id_prospect),
    INDEX idx_id_chatInput (id_chatInput)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
