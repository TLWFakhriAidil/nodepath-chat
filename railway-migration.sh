#!/bin/bash
# Railway Migration Script - Updates ai_whatsapp_nodepath schema in production
# This script removes deprecated columns and adds new schema columns during Railway deployment

echo "🚀 Starting Railway production migration..."

# Check if MYSQL_URI is set
if [ -z "$MYSQL_URI" ]; then
    echo "❌ Error: MYSQL_URI environment variable is not set"
    exit 1
fi

echo "✅ MYSQL_URI found, proceeding with migration..."

# Create temporary migration script
cat > /tmp/migration.sql << 'EOF'
-- Production Migration: Update ai_whatsapp_nodepath schema
-- This script removes deprecated columns and adds new schema columns

-- Drop deprecated columns
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ai_whatsapp_nodepath' AND COLUMN_NAME = 'jam');
SET @sql = IF(@col_exists > 0, 'ALTER TABLE ai_whatsapp_nodepath DROP COLUMN jam', 'SELECT "Column jam already removed" AS message');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ai_whatsapp_nodepath' AND COLUMN_NAME = 'conv_stage');
SET @sql = IF(@col_exists > 0, 'ALTER TABLE ai_whatsapp_nodepath DROP COLUMN conv_stage', 'SELECT "Column conv_stage already removed" AS message');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ai_whatsapp_nodepath' AND COLUMN_NAME = 'variables');
SET @sql = IF(@col_exists > 0, 'ALTER TABLE ai_whatsapp_nodepath DROP COLUMN variables', 'SELECT "Column variables already removed" AS message');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ai_whatsapp_nodepath' AND COLUMN_NAME = 'catatan_staff');
SET @sql = IF(@col_exists > 0, 'ALTER TABLE ai_whatsapp_nodepath DROP COLUMN catatan_staff', 'SELECT "Column catatan_staff already removed" AS message');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ai_whatsapp_nodepath' AND COLUMN_NAME = 'data_image');
SET @sql = IF(@col_exists > 0, 'ALTER TABLE ai_whatsapp_nodepath DROP COLUMN data_image', 'SELECT "Column data_image already removed" AS message');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ai_whatsapp_nodepath' AND COLUMN_NAME = 'current_node');
SET @sql = IF(@col_exists > 0, 'ALTER TABLE ai_whatsapp_nodepath DROP COLUMN current_node', 'SELECT "Column current_node already removed" AS message');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ai_whatsapp_nodepath' AND COLUMN_NAME = 'bot_balas');
SET @sql = IF(@col_exists > 0, 'ALTER TABLE ai_whatsapp_nodepath DROP COLUMN bot_balas', 'SELECT "Column bot_balas already removed" AS message');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Add new schema columns
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ai_whatsapp_nodepath' AND COLUMN_NAME = 'flow_reference');
SET @sql = IF(@col_exists = 0, 'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN flow_reference VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT "Reference to chatbot flow being executed"', 'SELECT "Column flow_reference already exists" AS message');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ai_whatsapp_nodepath' AND COLUMN_NAME = 'execution_id');
SET @sql = IF(@col_exists = 0, 'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN execution_id VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT "Unique execution identifier"', 'SELECT "Column execution_id already exists" AS message');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ai_whatsapp_nodepath' AND COLUMN_NAME = 'prospect_name');
SET @sql = IF(@col_exists = 0, 'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN prospect_name VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL', 'SELECT "Column prospect_name already exists" AS message');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ai_whatsapp_nodepath' AND COLUMN_NAME = 'execution_status');
SET @sql = IF(@col_exists = 0, 'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN execution_status ENUM("active","completed","failed") COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT "Flow execution status"', 'SELECT "Column execution_status already exists" AS message');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ai_whatsapp_nodepath' AND COLUMN_NAME = 'flow_id');
SET @sql = IF(@col_exists = 0, 'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN flow_id VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT "ID of the current chatbot flow being executed"', 'SELECT "Column flow_id already exists" AS message');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ai_whatsapp_nodepath' AND COLUMN_NAME = 'current_node_id');
SET @sql = IF(@col_exists = 0, 'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN current_node_id VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT "Current node ID in the chatbot flow"', 'SELECT "Column current_node_id already exists" AS message');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ai_whatsapp_nodepath' AND COLUMN_NAME = 'last_node_id');
SET @sql = IF(@col_exists = 0, 'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN last_node_id VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT "Previous node ID for flow tracking"', 'SELECT "Column last_node_id already exists" AS message');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ai_whatsapp_nodepath' AND COLUMN_NAME = 'waiting_for_reply');
SET @sql = IF(@col_exists = 0, 'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN waiting_for_reply TINYINT(1) DEFAULT 0 COMMENT "1 = waiting for user reply, 0 = not waiting"', 'SELECT "Column waiting_for_reply already exists" AS message');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Verify schema update
SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ai_whatsapp_nodepath' ORDER BY ORDINAL_POSITION;
EOF

echo "📝 Migration script created, executing..."

# Build and run the migration utility
go build -o /tmp/migrate fix_production_schema.go
if [ $? -eq 0 ]; then
    echo "✅ Migration utility built successfully"
    /tmp/migrate
    if [ $? -eq 0 ]; then
        echo "🎉 Production schema migration completed successfully!"
    else
        echo "❌ Migration failed, but continuing deployment..."
    fi
else
    echo "❌ Failed to build migration utility, skipping migration..."
fi

echo "🚀 Railway migration script completed"