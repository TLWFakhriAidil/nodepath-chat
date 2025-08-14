-- Migration: Create broadcast_messages table for high-performance broadcast system
-- This table is optimized for high throughput and supports campaigns, sequences, and bulk operations

CREATE TABLE IF NOT EXISTS broadcast_messages (
    id VARCHAR(36) PRIMARY KEY COMMENT 'Unique message identifier (UUID)',
    user_id VARCHAR(36) NOT NULL COMMENT 'User who owns this message',
    device_id VARCHAR(36) NOT NULL COMMENT 'Device that will send this message',
    
    -- Campaign and Sequence support
    campaign_id INT NULL COMMENT 'Campaign ID if this is a campaign message',
    sequence_id VARCHAR(36) NULL COMMENT 'Sequence ID if this is a sequence message',
    sequence_stepid VARCHAR(36) NULL COMMENT 'Sequence step ID for ordering and duplicate prevention',
    
    -- Recipient information
    recipient_phone VARCHAR(20) NOT NULL COMMENT 'Recipient phone number in international format',
    recipient_name VARCHAR(255) NULL COMMENT 'Recipient display name',
    
    -- Message content
    message_type ENUM('text', 'image', 'video', 'audio', 'document') DEFAULT 'text' COMMENT 'Type of message content',
    content TEXT NOT NULL COMMENT 'Message content (text or caption for media)',
    media_url VARCHAR(500) NULL COMMENT 'URL to media file if applicable',
    
    -- Processing status
    status ENUM('pending', 'queued', 'processing', 'sent', 'failed', 'cancelled') DEFAULT 'pending' COMMENT 'Current message status',
    
    -- Scheduling and timing
    scheduled_at TIMESTAMP NULL COMMENT 'When this message should be sent (NULL = immediate)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'When this message was created',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last status update time',
    
    -- Grouping and ordering
    group_id VARCHAR(36) NULL COMMENT 'Group ID for batch operations',
    group_order INT DEFAULT 0 COMMENT 'Order within group for sequential sending',
    
    -- Error handling
    error_message TEXT NULL COMMENT 'Error details if message failed',
    
    -- Performance indexes for high-throughput operations
    INDEX idx_device_status_scheduled (device_id, status, scheduled_at) COMMENT 'Primary query index for worker processing',
    INDEX idx_campaign_status (campaign_id, status) COMMENT 'Campaign management queries',
    INDEX idx_sequence_status (sequence_id, status) COMMENT 'Sequence management queries',
    INDEX idx_sequence_step_duplicate (sequence_stepid, recipient_phone, device_id, status) COMMENT 'Sequence duplicate prevention',
    INDEX idx_campaign_duplicate (campaign_id, recipient_phone, device_id, status) COMMENT 'Campaign duplicate prevention',
    INDEX idx_status_created (status, created_at) COMMENT 'Cleanup and monitoring queries',
    INDEX idx_user_device (user_id, device_id) COMMENT 'User device management',
    INDEX idx_updated_status (updated_at, status) COMMENT 'Stuck message detection'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='High-performance broadcast message queue';

-- Create a separate table for tracking device worker status (optional but recommended)
CREATE TABLE IF NOT EXISTS device_workers (
    device_id VARCHAR(36) PRIMARY KEY COMMENT 'Device identifier',
    status ENUM('idle', 'running', 'paused', 'error', 'stopped') DEFAULT 'idle' COMMENT 'Worker status',
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last worker activity',
    messages_processed BIGINT DEFAULT 0 COMMENT 'Total messages processed by this worker',
    messages_failed BIGINT DEFAULT 0 COMMENT 'Total messages failed by this worker',
    queue_size INT DEFAULT 0 COMMENT 'Current queue size for this worker',
    min_delay INT DEFAULT 5 COMMENT 'Minimum delay between messages (seconds)',
    max_delay INT DEFAULT 15 COMMENT 'Maximum delay between messages (seconds)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'When worker was first created',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last status update',
    
    INDEX idx_status_activity (status, last_activity) COMMENT 'Worker health monitoring'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Device worker status tracking';

-- Create a table for system metrics (optional but useful for monitoring)
CREATE TABLE IF NOT EXISTS broadcast_metrics (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    metric_name VARCHAR(100) NOT NULL COMMENT 'Name of the metric',
    metric_value DECIMAL(15,4) NOT NULL COMMENT 'Numeric value of the metric',
    metric_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'When this metric was recorded',
    device_id VARCHAR(36) NULL COMMENT 'Device ID if metric is device-specific',
    
    INDEX idx_metric_timestamp (metric_name, metric_timestamp) COMMENT 'Time-series queries',
    INDEX idx_device_metric (device_id, metric_name, metric_timestamp) COMMENT 'Device-specific metrics'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='System performance metrics storage';

-- Insert initial configuration data
INSERT INTO device_workers (device_id, status, min_delay, max_delay) 
VALUES ('default', 'idle', 5, 15) 
ON DUPLICATE KEY UPDATE 
    min_delay = VALUES(min_delay),
    max_delay = VALUES(max_delay);

-- Create a view for easy monitoring of queue status
CREATE OR REPLACE VIEW broadcast_queue_summary AS
SELECT 
    device_id,
    COUNT(*) as total_messages,
    SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending_count,
    SUM(CASE WHEN status = 'queued' THEN 1 ELSE 0 END) as queued_count,
    SUM(CASE WHEN status = 'processing' THEN 1 ELSE 0 END) as processing_count,
    SUM(CASE WHEN status = 'sent' THEN 1 ELSE 0 END) as sent_count,
    SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed_count,
    SUM(CASE WHEN status = 'cancelled' THEN 1 ELSE 0 END) as cancelled_count,
    MIN(created_at) as oldest_message,
    MAX(updated_at) as latest_activity
FROM broadcast_messages 
WHERE created_at > DATE_SUB(NOW(), INTERVAL 24 HOUR)
GROUP BY device_id;

-- Create a view for stuck message detection
CREATE OR REPLACE VIEW stuck_messages AS
SELECT 
    id,
    device_id,
    status,
    recipient_phone,
    created_at,
    updated_at,
    TIMESTAMPDIFF(MINUTE, updated_at, NOW()) as minutes_stuck
FROM broadcast_messages 
WHERE 
    (
        (status = 'queued' AND updated_at < DATE_SUB(NOW(), INTERVAL 10 MINUTE))
        OR 
        (status = 'processing' AND updated_at < DATE_SUB(NOW(), INTERVAL 15 MINUTE))
        OR
        (status IN ('pending', 'queued', 'processing') AND created_at < DATE_SUB(NOW(), INTERVAL 12 HOUR))
    )
ORDER BY updated_at ASC;

-- Performance optimization: Create a procedure for bulk message insertion
DELIMITER //

CREATE PROCEDURE IF NOT EXISTS BulkInsertMessages(
    IN message_data JSON
)
BEGIN
    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;
    
    START TRANSACTION;
    
    INSERT INTO broadcast_messages (
        id, user_id, device_id, campaign_id, sequence_id, sequence_stepid,
        recipient_phone, recipient_name, message_type, content, media_url,
        status, scheduled_at, group_id, group_order
    )
    SELECT 
        JSON_UNQUOTE(JSON_EXTRACT(value, '$.id')),
        JSON_UNQUOTE(JSON_EXTRACT(value, '$.user_id')),
        JSON_UNQUOTE(JSON_EXTRACT(value, '$.device_id')),
        JSON_EXTRACT(value, '$.campaign_id'),
        JSON_UNQUOTE(JSON_EXTRACT(value, '$.sequence_id')),
        JSON_UNQUOTE(JSON_EXTRACT(value, '$.sequence_stepid')),
        JSON_UNQUOTE(JSON_EXTRACT(value, '$.recipient_phone')),
        JSON_UNQUOTE(JSON_EXTRACT(value, '$.recipient_name')),
        JSON_UNQUOTE(JSON_EXTRACT(value, '$.message_type')),
        JSON_UNQUOTE(JSON_EXTRACT(value, '$.content')),
        JSON_UNQUOTE(JSON_EXTRACT(value, '$.media_url')),
        COALESCE(JSON_UNQUOTE(JSON_EXTRACT(value, '$.status')), 'pending'),
        JSON_UNQUOTE(JSON_EXTRACT(value, '$.scheduled_at')),
        JSON_UNQUOTE(JSON_EXTRACT(value, '$.group_id')),
        COALESCE(JSON_EXTRACT(value, '$.group_order'), 0)
    FROM JSON_TABLE(
        message_data,
        '$[*]' COLUMNS (
            value JSON PATH '$'
        )
    ) AS jt;
    
    COMMIT;
END //

DELIMITER ;

-- Create a procedure for queue cleanup
DELIMITER //

CREATE PROCEDURE IF NOT EXISTS CleanupStuckMessages()
BEGIN
    DECLARE stuck_queued_count INT DEFAULT 0;
    DECLARE stuck_processing_count INT DEFAULT 0;
    DECLARE timeout_count INT DEFAULT 0;
    
    -- Reset stuck queued messages (stuck for more than 10 minutes)
    UPDATE broadcast_messages 
    SET status = 'pending', 
        updated_at = NOW(),
        error_message = 'Reset from stuck queued status'
    WHERE status = 'queued' 
      AND updated_at < DATE_SUB(NOW(), INTERVAL 10 MINUTE);
    
    SET stuck_queued_count = ROW_COUNT();
    
    -- Reset stuck processing messages (stuck for more than 15 minutes)
    UPDATE broadcast_messages 
    SET status = 'pending', 
        updated_at = NOW(),
        error_message = 'Reset from stuck processing status'
    WHERE status = 'processing' 
      AND updated_at < DATE_SUB(NOW(), INTERVAL 15 MINUTE);
    
    SET stuck_processing_count = ROW_COUNT();
    
    -- Mark old messages as failed (older than 12 hours)
    UPDATE broadcast_messages 
    SET status = 'failed', 
        updated_at = NOW(),
        error_message = 'Message timeout - exceeded maximum processing time'
    WHERE status IN ('pending', 'queued', 'processing') 
      AND created_at < DATE_SUB(NOW(), INTERVAL 12 HOUR);
    
    SET timeout_count = ROW_COUNT();
    
    -- Log cleanup results
    INSERT INTO broadcast_metrics (metric_name, metric_value, metric_timestamp)
    VALUES 
        ('cleanup_stuck_queued', stuck_queued_count, NOW()),
        ('cleanup_stuck_processing', stuck_processing_count, NOW()),
        ('cleanup_timeout', timeout_count, NOW());
        
END //

DELIMITER ;

-- Create an event to run cleanup automatically every 5 minutes
-- Note: This requires the event scheduler to be enabled (SET GLOBAL event_scheduler = ON;)
CREATE EVENT IF NOT EXISTS cleanup_stuck_messages_event
ON SCHEDULE EVERY 5 MINUTE
STARTS CURRENT_TIMESTAMP
DO
  CALL CleanupStuckMessages();

-- Performance tips and notes:
-- 1. Ensure MySQL has sufficient innodb_buffer_pool_size (recommended: 70-80% of available RAM)
-- 2. Set innodb_log_file_size appropriately for high write loads
-- 3. Consider partitioning the broadcast_messages table by date for very high volumes
-- 4. Monitor slow query log and optimize queries as needed
-- 5. Use connection pooling in your application
-- 6. Consider read replicas for reporting queries

-- Example partitioning (uncomment if needed for very high volumes):
/*
ALTER TABLE broadcast_messages 
PARTITION BY RANGE (TO_DAYS(created_at)) (
    PARTITION p_2024_01 VALUES LESS THAN (TO_DAYS('2024-02-01')),
    PARTITION p_2024_02 VALUES LESS THAN (TO_DAYS('2024-03-01')),
    PARTITION p_2024_03 VALUES LESS THAN (TO_DAYS('2024-04-01')),
    PARTITION p_future VALUES LESS THAN MAXVALUE
);
*/