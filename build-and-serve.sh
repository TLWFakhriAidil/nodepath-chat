#!/bin/bash

# Build and Serve Script - Matches Railway Deployment
# This script ensures your local environment matches Railway's production setup

echo "🚀 Building frontend for production..."
npm run build

if [ $? -eq 0 ]; then
    echo "✅ Frontend build completed successfully"
    echo "🔄 Starting Go backend server (matches Railway setup)..."
    go run cmd/server/main.go
else
    echo "❌ Frontend build failed"
    exit 1
fi