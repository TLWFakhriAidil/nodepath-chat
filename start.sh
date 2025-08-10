#!/bin/bash

# Display environment information
echo "Starting NodePath Chat Go Server..."
echo "Go version: $(go version)"

# Set PORT environment variable for Railway
export PORT=${PORT:-8080}
echo "Server will start on port: $PORT"

# Build and run the Go application
echo "Building Go application..."
go build -o server cmd/server/main.go

if [ $? -eq 0 ]; then
    echo "Build successful. Starting server..."
    ./server
else
    echo "Build failed. Trying to run directly..."
    go run cmd/server/main.go
fi