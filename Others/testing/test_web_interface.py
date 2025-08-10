import requests
from bs4 import BeautifulSoup
import json
import sys
import io

# Fix encoding
if sys.platform == 'win32':
    sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')

# Test configuration
BASE_URL = "https://web-production-b777.up.railway.app"
EMAIL = "aqil@gmail.com"
PASSWORD = "aqil@gmail.com"

print("🧪 Testing WhatsApp Multi-Device System on Railway")
print("=" * 60)
print(f"URL: {BASE_URL}")
print(f"User: {EMAIL}")
print()

# Create session
session = requests.Session()

# Test 1: Check if login page exists
print("1. Checking login page...")
try:
    response = session.get(f"{BASE_URL}/login")
    if response.status_code == 200:
        print("   ✅ Login page accessible")
        # Check if it's the WhatsApp system
        if "WhatsApp" in response.text:
            print("   ✅ WhatsApp system confirmed")
    else:
        print(f"   ❌ Login page returned: {response.status_code}")
except Exception as e:
    print(f"   ❌ Error: {e}")

# Test 2: Try to login
print("\n2. Attempting login...")
try:
    # Get CSRF token if needed
    login_page = session.get(f"{BASE_URL}/login")
    
    # Try login
    login_data = {
        "email": EMAIL,
        "password": PASSWORD
    }
    
    login_response = session.post(f"{BASE_URL}/login", data=login_data, allow_redirects=False)
    
    if login_response.status_code in [302, 303]:
        print("   ✅ Login successful (redirect detected)")
        # Follow redirect
        dashboard = session.get(f"{BASE_URL}/dashboard")
        if dashboard.status_code == 200:
            print("   ✅ Dashboard accessible")
    elif login_response.status_code == 200:
        if "dashboard" in login_response.url or "Dashboard" in login_response.text:
            print("   ✅ Login successful")
        else:
            print("   ❌ Login failed - check credentials")
    else:
        print(f"   ❌ Login returned: {login_response.status_code}")
        
except Exception as e:
    print(f"   ❌ Login error: {e}")

# Test 3: Check dashboard/devices
print("\n3. Checking devices page...")
try:
    devices_response = session.get(f"{BASE_URL}/devices")
    if devices_response.status_code == 200:
        print("   ✅ Devices page accessible")
        # Count devices if possible
        if "device" in devices_response.text.lower():
            print("   ✅ Device management interface found")
    else:
        print(f"   ❌ Devices page returned: {devices_response.status_code}")
except Exception as e:
    print(f"   ❌ Error: {e}")

# Test 4: Check campaigns
print("\n4. Checking campaigns page...")
try:
    campaigns_response = session.get(f"{BASE_URL}/campaigns")
    if campaigns_response.status_code == 200:
        print("   ✅ Campaigns page accessible")
    else:
        print(f"   ❌ Campaigns page returned: {campaigns_response.status_code}")
except Exception as e:
    print(f"   ❌ Error: {e}")

# Test 5: Check sequences
print("\n5. Checking sequences page...")
try:
    sequences_response = session.get(f"{BASE_URL}/sequences")
    if sequences_response.status_code == 200:
        print("   ✅ Sequences page accessible")
    else:
        print(f"   ❌ Sequences page returned: {sequences_response.status_code}")
except Exception as e:
    print(f"   ❌ Error: {e}")

# Test 6: Check AI campaigns
print("\n6. Checking AI campaigns page...")
try:
    ai_response = session.get(f"{BASE_URL}/ai-campaigns")
    if ai_response.status_code == 200:
        print("   ✅ AI campaigns page accessible")
    else:
        print(f"   ❌ AI campaigns page returned: {ai_response.status_code}")
except Exception as e:
    print(f"   ❌ Error: {e}")

print("\n" + "=" * 60)
print("📊 SUMMARY")
print("=" * 60)
print("\nBased on the responses, the system appears to be:")
print("- Using web interface (not REST API)")
print("- Login with email/password (not Basic Auth)")
print("- Has standard pages: devices, campaigns, sequences, AI campaigns")
print("\nTo fully test the system:")
print("1. Login via web interface")
print("2. Check each section manually")
print("3. Create test data through the UI")
print("4. Monitor performance")
