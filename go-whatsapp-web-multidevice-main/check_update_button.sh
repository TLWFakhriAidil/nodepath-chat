#!/bin/bash
echo "Checking if UPDATE button changes are in the build..."

# Check if the UPDATE button text exists in the HTML files
echo "Checking sequences.html:"
grep -n "arrow-repeat" src/views/sequences.html | head -5

echo -e "\nChecking dashboard.html:"
grep -n "arrow-repeat" src/views/dashboard.html | head -5

echo -e "\nChecking API endpoint:"
grep -n "update-pending-messages" src/ui/rest/app.go | head -5

echo -e "\nDone!"
