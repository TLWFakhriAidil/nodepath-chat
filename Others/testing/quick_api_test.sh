#!/bin/bash
# Quick API test script

echo "🧪 Quick WhatsApp API Test"
echo "=========================="
echo ""

# Get URL from user
read -p "Enter your Railway URL (e.g., https://your-app.up.railway.app): " RAILWAY_URL
read -p "Enter username (default: admin): " AUTH_USER
read -p "Enter password (default: changeme123): " AUTH_PASS

AUTH_USER=${AUTH_USER:-admin}
AUTH_PASS=${AUTH_PASS:-changeme123}

echo ""
echo "Testing: $RAILWAY_URL"
echo ""

# Test 1: Basic connection
echo "1. Testing basic connection..."
curl -s -o /dev/null -w "   HTTP Status: %{http_code}\n" \
  -u "$AUTH_USER:$AUTH_PASS" \
  "$RAILWAY_URL/api/v1/check-server"

# Test 2: Get devices
echo ""
echo "2. Getting devices..."
DEVICES=$(curl -s -u "$AUTH_USER:$AUTH_PASS" "$RAILWAY_URL/api/v1/devices")
if [ $? -eq 0 ]; then
  echo "$DEVICES" | python -m json.tool 2>/dev/null | head -20
else
  echo "   Failed to get devices"
fi

# Test 3: Get campaigns
echo ""
echo "3. Getting campaigns..."
CAMPAIGNS=$(curl -s -u "$AUTH_USER:$AUTH_PASS" "$RAILWAY_URL/api/v1/campaigns")
if [ $? -eq 0 ]; then
  echo "$CAMPAIGNS" | python -m json.tool 2>/dev/null | head -20
else
  echo "   Failed to get campaigns"
fi

echo ""
echo "Test complete!"
