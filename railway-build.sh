#!/bin/bash

# Display Node.js and npm versions
echo "Node.js version:"
node -v
echo "npm version:"
npm -v

# Install dependencies
echo "Installing dependencies..."
npm install

# Build the application
echo "Building the application..."
npm run build

# Check build status
if [ $? -eq 0 ]; then
  echo "Build completed successfully"
else
  echo "Build failed with exit code $?"
  exit 1
fi