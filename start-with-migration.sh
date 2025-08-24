#!/bin/bash

# Railway startup script with database migration
# This script runs the migration before starting the main server

echo "🚀 Starting Railway deployment with database migration..."

# Check if MYSQL_URI is available
if [ -z "$MYSQL_URI" ]; then
    echo "⚠️ MYSQL_URI not found, skipping migration"
else
    echo "📡 MYSQL_URI found, running comprehensive migration..."
    
    # Run the Railway migration runner
    echo "🔄 Executing comprehensive database migration..."
    /app/railway_migration_runner
    
    migration_exit_code=$?
    if [ $migration_exit_code -eq 0 ]; then
        echo "✅ Comprehensive migration completed successfully"
    else
        echo "⚠️ Migration failed with exit code $migration_exit_code, but continuing..."
        echo "ℹ️ Server will start anyway to maintain service availability"
    fi
fi

echo "🚀 Starting main server..."
# Start the main application
exec /app/server