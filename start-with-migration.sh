#!/bin/sh
# Railway Startup Script - Runs database migration before starting server
# This ensures the 'jam' column exists before processing webhooks

echo "🚀 Starting Railway application with database migration..."

# Check if MYSQL_URI is set
if [ -z "$MYSQL_URI" ]; then
    echo "⚠️  Warning: MYSQL_URI environment variable is not set"
    echo "🚀 Starting server without migration..."
    exec /app/server
fi

echo "✅ MYSQL_URI found, running database migration..."

# Run the migration utility
echo "🔧 Executing database schema migration..."
/app/migrate

if [ $? -eq 0 ]; then
    echo "🎉 Database migration completed successfully!"
else
    echo "⚠️  Migration failed or skipped, continuing with server startup..."
fi

echo "🚀 Starting main application server..."

# Start the main application
exec /app/server