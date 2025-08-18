#!/bin/bash
# Railway Migration Script - Fixes missing 'jam' column in production
# This script runs during Railway deployment to ensure database schema is correct

echo "🚀 Starting Railway production migration..."

# Check if MYSQL_URI is set
if [ -z "$MYSQL_URI" ]; then
    echo "❌ Error: MYSQL_URI environment variable is not set"
    exit 1
fi

echo "✅ MYSQL_URI found, proceeding with migration..."

# Create temporary migration script
cat > /tmp/migration.sql << 'EOF'
-- Production Migration: Add missing 'jam' column to ai_whatsapp_nodepath table
-- This script fixes the "Unknown column 'jam' in 'field list'" error

-- Check if the column exists before adding it
SET @column_exists = (
    SELECT COUNT(*)
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'ai_whatsapp_nodepath'
    AND COLUMN_NAME = 'jam'
);

-- Add the jam column if it doesn't exist
SET @sql = IF(@column_exists = 0,
    'ALTER TABLE ai_whatsapp_nodepath ADD COLUMN jam VARCHAR(255) DEFAULT NULL',
    'SELECT "Column jam already exists" AS message'
);

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Verify the column was added
SELECT 
    COLUMN_NAME,
    DATA_TYPE,
    IS_NULLABLE,
    COLUMN_DEFAULT
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
AND TABLE_NAME = 'ai_whatsapp_nodepath'
AND COLUMN_NAME = 'jam';
EOF

echo "📝 Migration script created, executing..."

# Build and run the migration utility
go build -o /tmp/migrate fix_production_schema.go
if [ $? -eq 0 ]; then
    echo "✅ Migration utility built successfully"
    /tmp/migrate
    if [ $? -eq 0 ]; then
        echo "🎉 Production migration completed successfully!"
    else
        echo "❌ Migration failed, but continuing deployment..."
    fi
else
    echo "❌ Failed to build migration utility, skipping migration..."
fi

echo "🚀 Railway migration script completed"