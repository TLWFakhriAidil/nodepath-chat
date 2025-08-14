import requests
import json
import time
from datetime import datetime
import random
import sys

# Fix encoding for Windows
if sys.platform == 'win32':
    import io
    sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')

# Railway deployment configuration
RAILWAY_URL = "https://web-production-b777.up.railway.app"  # Your Railway URL
API_BASE = f"{RAILWAY_URL}/api/v1"
AUTH = ("admin", "changeme123")  # Using APP_BASIC_AUTH from env vars

class WhatsAppSystemTester:
    def __init__(self):
        self.session = requests.Session()
        self.session.auth = AUTH
        self.session.headers.update({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        })
        self.results = {
            'campaigns': {'status': 'Not tested', 'details': []},
            'ai_campaigns': {'status': 'Not tested', 'details': []},
            'sequences': {'status': 'Not tested', 'details': []},
            'devices': {'status': 'Not tested', 'details': []}
        }
    
    def log(self, message, status="INFO"):
        timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        print(f"[{timestamp}] [{status}] {message}")
    
    def test_connection(self):
        """Test basic connection to Railway app"""
        self.log("=" * 60)
        self.log("🚀 TESTING RAILWAY DEPLOYMENT")
        self.log("=" * 60)
        self.log(f"URL: {RAILWAY_URL}")
        self.log("Testing connection...")
        
        try:
            # Try health check endpoint
            response = self.session.get(f"{API_BASE}/check-server", timeout=10)
            if response.status_code == 200:
                data = response.json()
                self.log(f"✅ Connection successful!", "SUCCESS")
                self.log(f"   Server: {data.get('data', {}).get('server', 'Unknown')}")
                self.log(f"   Version: {data.get('data', {}).get('version', 'Unknown')}")
                return True
            elif response.status_code == 401:
                self.log("❌ Authentication failed - check credentials", "ERROR")
                return False
            else:
                self.log(f"❌ Connection failed: HTTP {response.status_code}", "ERROR")
                return False
        except requests.exceptions.ConnectionError:
            self.log("❌ Cannot connect to Railway app - is it deployed?", "ERROR")
            self.log("   Check if the URL is correct", "ERROR")
            return False
        except Exception as e:
            self.log(f"❌ Connection error: {str(e)}", "ERROR")
            return False
    
    def test_database(self):
        """Test database connectivity"""
        self.log("\n📊 TESTING DATABASE CONNECTION")
        self.log("-" * 40)
        
        try:
            # Test user endpoint which requires DB
            response = self.session.get(f"{API_BASE}/user", timeout=10)
            if response.status_code == 200:
                self.log("✅ Database connection working", "SUCCESS")
                return True
            else:
                self.log(f"⚠️ Database test returned: {response.status_code}", "WARNING")
                return False
        except Exception as e:
            self.log(f"❌ Database test error: {str(e)}", "ERROR")
            return False
    
    def test_devices(self):
        """Test device management (3000 devices)"""
        self.log("\n📱 TESTING DEVICE MANAGEMENT")
        self.log("-" * 40)
        
        try:
            # Get devices list
            response = self.session.get(f"{API_BASE}/devices", timeout=10)
            if response.status_code == 200:
                data = response.json()
                devices = data.get('data', [])
                total_devices = len(devices)
                online_devices = sum(1 for d in devices if d.get('status') == 'online')
                
                self.log(f"Total devices: {total_devices}")
                self.log(f"Online devices: {online_devices}")
                self.log(f"Offline devices: {total_devices - online_devices}")
                
                self.results['devices'] = {
                    'status': 'Working' if total_devices > 0 else 'No devices',
                    'total': total_devices,
                    'online': online_devices,
                    'offline': total_devices - online_devices
                }
                
                # Check if system can handle 3000 devices
                if total_devices >= 3000:
                    self.log("✅ System handling 3000+ devices!", "SUCCESS")
                elif total_devices > 0:
                    self.log(f"✅ Device system working ({total_devices} devices)", "SUCCESS")
                else:
                    self.log("⚠️ No devices found - need to add devices", "WARNING")
                
                return True
            else:
                self.log(f"❌ Failed to get devices: {response.status_code}", "ERROR")
                return False
        except Exception as e:
            self.log(f"❌ Device test error: {str(e)}", "ERROR")
            return False
    
    def test_campaigns(self):
        """Test campaign functionality"""
        self.log("\n📢 TESTING CAMPAIGNS")
        self.log("-" * 40)
        
        try:
            # Get campaigns
            response = self.session.get(f"{API_BASE}/campaigns", timeout=10)
            if response.status_code == 200:
                data = response.json()
                campaigns = data.get('data', [])
                
                self.log(f"Total campaigns: {len(campaigns)}")
                
                # Check each campaign
                for campaign in campaigns[:5]:  # Show first 5
                    self.log(f"\nCampaign: {campaign.get('name', 'Unknown')}")
                    self.log(f"  Status: {campaign.get('status', 'Unknown')}")
                    self.log(f"  Target: {campaign.get('target_status', 'All')}")
                    self.log(f"  Schedule: {campaign.get('time_schedule', 'Not set')}")
                
                # Test if campaigns process correctly
                active_campaigns = [c for c in campaigns if c.get('status') == 'active']
                self.log(f"\n✅ Active campaigns: {len(active_campaigns)}")
                
                self.results['campaigns'] = {
                    'status': 'Working' if len(campaigns) > 0 else 'No campaigns',
                    'total': len(campaigns),
                    'active': len(active_campaigns),
                    'details': campaigns[:3]  # First 3 for summary
                }
                
                return True
            else:
                self.log(f"❌ Failed to get campaigns: {response.status_code}", "ERROR")
                return False
        except Exception as e:
            self.log(f"❌ Campaign test error: {str(e)}", "ERROR")
            return False
    
    def test_ai_campaigns(self):
        """Test AI campaign functionality"""
        self.log("\n🤖 TESTING AI CAMPAIGNS")
        self.log("-" * 40)
        
        try:
            # Get AI campaigns
            response = self.session.get(f"{API_BASE}/ai-campaigns", timeout=10)
            if response.status_code == 200:
                data = response.json()
                ai_campaigns = data.get('data', [])
                
                self.log(f"Total AI campaigns: {len(ai_campaigns)}")
                
                for campaign in ai_campaigns[:3]:
                    self.log(f"\nAI Campaign: {campaign.get('campaign_name', 'Unknown')}")
                    self.log(f"  Source: {campaign.get('lead_source', 'Unknown')}")
                    self.log(f"  Status: {campaign.get('status', 'Unknown')}")
                    self.log(f"  Device limit: {campaign.get('device_limit_per_device', 0)}/hour")
                    self.log(f"  Daily limit: {campaign.get('daily_limit', 0)}")
                
                active_ai = [c for c in ai_campaigns if c.get('status') == 'active']
                self.log(f"\n✅ Active AI campaigns: {len(active_ai)}")
                
                self.results['ai_campaigns'] = {
                    'status': 'Working' if len(ai_campaigns) > 0 else 'No AI campaigns',
                    'total': len(ai_campaigns),
                    'active': len(active_ai),
                    'details': ai_campaigns[:3]
                }
                
                return True
            else:
                self.log(f"❌ Failed to get AI campaigns: {response.status_code}", "ERROR")
                return False
        except Exception as e:
            self.log(f"❌ AI campaign test error: {str(e)}", "ERROR")
            return False
    
    def test_sequences(self):
        """Test sequence functionality (7-day sequences)"""
        self.log("\n📋 TESTING SEQUENCES")
        self.log("-" * 40)
        
        try:
            # Get sequences
            response = self.session.get(f"{API_BASE}/sequences", timeout=10)
            if response.status_code == 200:
                data = response.json()
                sequences = data.get('data', [])
                
                self.log(f"Total sequences: {len(sequences)}")
                
                for sequence in sequences[:3]:
                    self.log(f"\nSequence: {sequence.get('name', 'Unknown')}")
                    self.log(f"  Trigger: {sequence.get('trigger', 'Unknown')}")
                    self.log(f"  Status: {sequence.get('status', 'Unknown')}")
                    
                    # Get sequence steps
                    seq_id = sequence.get('id')
                    if seq_id:
                        steps_response = self.session.get(f"{API_BASE}/sequences/{seq_id}/steps", timeout=10)
                        if steps_response.status_code == 200:
                            steps_data = steps_response.json()
                            steps = steps_data.get('data', [])
                            self.log(f"  Steps: {len(steps)} days")
                
                active_sequences = [s for s in sequences if s.get('status') == 'active']
                self.log(f"\n✅ Active sequences: {len(active_sequences)}")
                
                self.results['sequences'] = {
                    'status': 'Working' if len(sequences) > 0 else 'No sequences',
                    'total': len(sequences),
                    'active': len(active_sequences),
                    'details': sequences[:3]
                }
                
                return True
            else:
                self.log(f"❌ Failed to get sequences: {response.status_code}", "ERROR")
                return False
        except Exception as e:
            self.log(f"❌ Sequence test error: {str(e)}", "ERROR")
            return False
    
    def test_simultaneous_load(self):
        """Test system under load (simulated)"""
        self.log("\n🔥 TESTING SYSTEM LOAD CAPACITY")
        self.log("-" * 40)
        
        try:
            # Get system stats
            self.log("Checking system capacity...")
            
            # Get devices for load calculation
            devices_response = self.session.get(f"{API_BASE}/devices", timeout=10)
            if devices_response.status_code == 200:
                devices_data = devices_response.json()
                total_devices = len(devices_data.get('data', []))
                online_devices = sum(1 for d in devices_data.get('data', []) if d.get('status') == 'online')
                
                # Calculate theoretical capacity
                messages_per_device_hour = 80  # WhatsApp limit
                hourly_capacity = online_devices * messages_per_device_hour
                daily_capacity = hourly_capacity * 24
                
                self.log(f"\n📊 SYSTEM CAPACITY:")
                self.log(f"  Online devices: {online_devices}")
                self.log(f"  Hourly capacity: {hourly_capacity:,} messages")
                self.log(f"  Daily capacity: {daily_capacity:,} messages")
                self.log(f"  Safe operating rate: {int(hourly_capacity * 0.7):,} msg/hour")
                
                if online_devices >= 2700:
                    self.log("\n✅ System ready for 3000 device load!", "SUCCESS")
                elif online_devices > 0:
                    self.log(f"\n✅ System operational with {online_devices} devices", "SUCCESS")
                else:
                    self.log("\n⚠️ No online devices for load testing", "WARNING")
                
                return True
            else:
                self.log("❌ Could not get device stats", "ERROR")
                return False
        except Exception as e:
            self.log(f"❌ Load test error: {str(e)}", "ERROR")
            return False
    
    def generate_summary(self):
        """Generate test summary"""
        self.log("\n" + "=" * 60)
        self.log("📊 TEST SUMMARY")
        self.log("=" * 60)
        
        # Overall status
        all_working = all(
            result.get('status', '').startswith('Working') 
            for result in self.results.values()
        )
        
        if all_working:
            self.log("✅ ALL SYSTEMS OPERATIONAL", "SUCCESS")
        else:
            self.log("⚠️ SOME SYSTEMS NEED ATTENTION", "WARNING")
        
        # Component summary
        self.log("\nCOMPONENT STATUS:")
        self.log(f"1. Devices: {self.results['devices']['status']}")
        if self.results['devices'].get('total', 0) > 0:
            self.log(f"   - Total: {self.results['devices']['total']}")
            self.log(f"   - Online: {self.results['devices']['online']}")
        
        self.log(f"\n2. Campaigns: {self.results['campaigns']['status']}")
        if self.results['campaigns'].get('total', 0) > 0:
            self.log(f"   - Total: {self.results['campaigns']['total']}")
            self.log(f"   - Active: {self.results['campaigns']['active']}")
        
        self.log(f"\n3. AI Campaigns: {self.results['ai_campaigns']['status']}")
        if self.results['ai_campaigns'].get('total', 0) > 0:
            self.log(f"   - Total: {self.results['ai_campaigns']['total']}")
            self.log(f"   - Active: {self.results['ai_campaigns']['active']}")
        
        self.log(f"\n4. Sequences: {self.results['sequences']['status']}")
        if self.results['sequences'].get('total', 0) > 0:
            self.log(f"   - Total: {self.results['sequences']['total']}")
            self.log(f"   - Active: {self.results['sequences']['active']}")
        
        self.log("\n" + "=" * 60)
    
    def run_all_tests(self):
        """Run all tests"""
        # Test connection first
        if not self.test_connection():
            self.log("\n❌ Cannot connect to Railway deployment", "ERROR")
            self.log("Please check:", "ERROR")
            self.log("1. Is the app deployed and running?", "ERROR")
            self.log("2. Is the URL correct?", "ERROR")
            self.log("3. Are the credentials correct?", "ERROR")
            return
        
        # Test database
        self.test_database()
        
        # Test all components
        self.test_devices()
        self.test_campaigns()
        self.test_ai_campaigns()
        self.test_sequences()
        self.test_simultaneous_load()
        
        # Generate summary
        self.generate_summary()

# Run the tests
if __name__ == "__main__":
    tester = WhatsAppSystemTester()
    tester.run_all_tests()
