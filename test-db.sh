#!/bin/bash

# Check if dotenv is installed
if ! npm list dotenv | grep -q dotenv; then
  echo "Installing dotenv package..."
  npm install dotenv
fi

# Run the database connection test
echo "Running database connection test..."
node test-db-connection.js

# Check if the test was successful
if [ $? -ne 0 ]; then
  echo -e "\nDatabase connection test failed with exit code $?" >&2
  exit 1
else
  echo -e "\nDatabase connection test completed successfully"
fi