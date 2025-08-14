import requests
import json
import time
from datetime import datetime

# Configuration - Update these with your Railway app details
RAILWAY_URL = "https://your-app.up.railway.app"  # Replace with your Railway URL
API_BASE = f"{RAILWAY_URL}/api/v1"
AUTH = ("admin", "changeme123")  # Replace with your APP_BASIC_AUTH credentials

class RailwayTester:
    def __init__(self):
        self.session = requests.Session()
        self.session.auth = AUTH
        self.results = []
    
    def log(self, message, status="INFO"):
        timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        print(f"[{timestamp}] [{status}] {message}")
        self.results.append({
            "time": timestamp,
            "status": status,
            "message": message
        })
    
    def test_connection(self):
        """Test basic connection to Railway app"""
        self.log("Testing connection to Railway deployment...")
        try:
            response = self.session.get(f"{RAILWAY_URL}/health", timeout=10)
            if response.status_code == 200:
                self.log("✅ Connection successful!", "SUCCESS")
                return True
            else:
                self.log(f"❌ Connection failed: {response.status_code}", "ERROR")
                return False
        except Exception as e:
            self.log(f"❌ Connection error: {str(e)}", "ERROR")
            return False
    
    def test_auth(self):
        """Test authentication"""
        self.log("Testing authentication...")
        try:
            response = self.session.get(f"{API_BASE}/check-server", timeout=10)
            if response.status_code == 200:
                data = response.json()
                self.log(f"✅ Auth successful! Server: {data.get('data', {}).get('server', 'Unknown')}", "SUCCESS")
                return True
            elif response.status_code == 401:
                self.log("❌ Authentication failed - check APP_BASIC_AUTH", "ERROR")
                return False
            else:
                self.log(f"❌ Unexpected response: {response.status_code}", "ERROR")
                return False
        except Exception as e:
            self.log(f"❌ Auth test error: {str(e)}", "ERROR")
            return False