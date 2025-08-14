import requests
import json
import time
from datetime import datetime
import sys
import io

# Fix encoding
if sys.platform == 'win32':
    sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')

class WorkerProcessingTest:
    def __init__(self):
        self.base_url = "https://web-production-b777.up.railway.app"
        self.session = requests.Session()
        
    def log(self, message, level="INFO"):
        print(f"[{datetime.now().strftime('%H:%M:%S')}] [{level}] {message}")
        
    def check_logs(self):
        """Check Railway logs for worker activity"""
        self.log("=" * 60)
        self.log("🔍 CHECKING WORKER ACTIVITY IN RAILWAY LOGS")
        self.log("=" * 60)
        
        self.log("\n📋 WHAT TO LOOK FOR IN RAILWAY LOGS:")
        self.log("-" * 40)
        
        # Campaign Worker
        self.log("\n1️⃣ CAMPAIGN WORKER:")
        self.log("   Look for these log messages:")
        self.log("   ✓ 'Starting campaign broadcast worker...'")
        self.log("   ✓ 'Processing campaign: [campaign_name]'")
        self.log("   ✓ 'Using X devices for broadcast'")
        self.log("   ✓ 'Sent X messages in Y seconds'")
        self.log("   ✓ 'Campaign worker sleeping for 30 seconds'")
        
        # Sequence Worker
        self.log("\n2️⃣ SEQUENCE WORKER:")
        self.log("   Look for these log messages:")
        self.log("   ✓ 'Starting sequence processor...'")
        self.log("   ✓ 'Processing sequence triggers...'")
        self.log("   ✓ 'Found X contacts ready for next message'")
        self.log("   ✓ 'Sequence: [name] - Processing X leads'")
        self.log("   ✓ 'Sent sequence message to X contacts'")
        
        # AI Campaign Worker
        self.log("\n3️⃣ AI CAMPAIGN WORKER:")
        self.log("   Look for these log messages:")
        self.log("   ✓ 'Starting AI campaign processor...'")
        self.log("   ✓ 'Processing AI campaign: [campaign_name]'")
        self.log("   ✓ 'Distributing X leads across Y devices'")
        self.log("   ✓ 'Device [id] assigned X leads'")
        self.log("   ✓ 'AI campaign processed X leads'")
        
        # Device Status
        self.log("\n4️⃣ DEVICE STATUS:")
        self.log("   Look for these log messages:")
        self.log("   ✓ 'Device TestDevice0001 connected'")
        self.log("   ✓ 'Online devices: X/3000'")
        self.log("   ✓ 'Device [name] status: online'")
        
        # Performance Metrics
        self.log("\n5️⃣ PERFORMANCE METRICS:")
        self.log("   Look for these log messages:")
        self.log("   ✓ 'Message rate: X msg/sec'")
        self.log("   ✓ 'Database query time: Xms'")
        self.log("   ✓ 'Redis cache hit rate: X%'")
        self.log("   ✓ 'Memory usage: X MB'")
        
    def check_database_activity(self):
        """Generate SQL queries to check worker activity"""
        self.log("\n" + "=" * 60)
        self.log("🗄️ DATABASE QUERIES TO CHECK WORKER ACTIVITY")
        self.log("=" * 60)
        
        self.log("\nRun these queries in your PostgreSQL to check worker activity:")
        self.log("-" * 40)
        
        # Check recent broadcast messages
        self.log("\n-- 1. Check recent campaign broadcasts (last hour):")
        self.log("""
SELECT 
    COUNT(*) as total_messages,
    COUNT(DISTINCT device_id) as devices_used,
    COUNT(DISTINCT campaign_id) as campaigns_processed,
    MIN(created_at) as first_message,
    MAX(created_at) as last_message
FROM broadcast_messages
WHERE created_at > NOW() - INTERVAL '1 hour';
        """)
        
        # Check sequence processing
        self.log("\n-- 2. Check sequence processing (last hour):")
        self.log("""
SELECT 
    s.name as sequence_name,
    COUNT(sc.*) as contacts_processed,
    COUNT(CASE WHEN sc.status = 'sent' THEN 1 END) as sent,
    COUNT(CASE WHEN sc.status = 'failed' THEN 1 END) as failed,
    MAX(sc.updated_at) as last_processed
FROM sequences s
LEFT JOIN sequence_contacts sc ON sc.sequence_id = s.id
WHERE sc.updated_at > NOW() - INTERVAL '1 hour'
GROUP BY s.id, s.name
ORDER BY last_processed DESC;
        """)
        
        # Check AI campaign activity
        self.log("\n-- 3. Check AI campaign processing:")
        self.log("""
SELECT 
    ac.campaign_name,
    COUNT(acl.*) as leads_processed,
    COUNT(DISTINCT acl.device_id) as devices_used,
    MIN(acl.assigned_at) as first_assignment,
    MAX(acl.assigned_at) as last_assignment
FROM ai_campaigns ac
LEFT JOIN ai_campaign_leads acl ON acl.campaign_id = ac.id
WHERE acl.assigned_at > NOW() - INTERVAL '1 hour'
GROUP BY ac.id, ac.campaign_name;
        """)
        
        # Check device activity
        self.log("\n-- 4. Check device activity:")
        self.log("""
SELECT 
    status,
    COUNT(*) as device_count,
    COUNT(CASE WHEN last_seen > NOW() - INTERVAL '5 minutes' THEN 1 END) as active_recently
FROM user_devices
WHERE device_name LIKE 'TestDevice%'
GROUP BY status;
        """)
        
        # Check worker health
        self.log("\n-- 5. Check message processing rate:")
        self.log("""
-- Messages per minute for last hour
WITH minute_stats AS (
    SELECT 
        DATE_TRUNC('minute', created_at) as minute,
        COUNT(*) as messages_sent
    FROM broadcast_messages
    WHERE created_at > NOW() - INTERVAL '1 hour'
    GROUP BY DATE_TRUNC('minute', created_at)
)
SELECT 
    AVG(messages_sent) as avg_messages_per_minute,
    MAX(messages_sent) as peak_messages_per_minute,
    MIN(messages_sent) as min_messages_per_minute
FROM minute_stats;
        """)
        
    def generate_test_data_sql(self):
        """Generate SQL to create test data"""
        self.log("\n" + "=" * 60)
        self.log("💾 SQL TO CREATE TEST DATA FOR WORKER TESTING")
        self.log("=" * 60)
        
        self.log("\nIf no data exists, run this SQL to create test data:")
        self.log("-" * 40)
        
        self.log("""
-- Create test campaign that should process immediately
INSERT INTO campaigns (id, user_id, name, message, status, campaign_date, time_schedule, target_status, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    (SELECT id FROM users WHERE email = 'aqil@gmail.com'),
    'Worker Test Campaign - ' || NOW()::TEXT,
    'Testing worker processing at {time}. Device: {device}',
    'active',
    CURRENT_DATE,
    '00:00-23:59', -- Run all day
    'Active',
    NOW(),
    NOW()
);

-- Create test sequence that should trigger
INSERT INTO sequences (id, user_id, name, niche, trigger, status, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    (SELECT id FROM users WHERE email = 'aqil@gmail.com'),
    'Worker Test Sequence - ' || NOW()::TEXT,
    'test',
    'worker_test',
    'active',
    NOW(),
    NOW()
);

-- Add some test leads with the trigger
INSERT INTO leads (id, user_id, device_id, name, phone, status, trigger, created_at, updated_at)
SELECT 
    gen_random_uuid(),
    (SELECT id FROM users WHERE email = 'aqil@gmail.com'),
    (SELECT id FROM user_devices ORDER BY RANDOM() LIMIT 1),
    'WorkerTestLead' || generate_series,
    '60' || LPAD(FLOOR(RANDOM() * 999999999)::TEXT, 9, '0'),
    'Active',
    'worker_test',
    NOW(),
    NOW()
FROM generate_series(1, 100);

-- Create AI campaign for immediate processing
INSERT INTO ai_campaigns (
    id, user_id, campaign_name, lead_source, lead_status,
    min_delay, max_delay, device_limit_per_device,
    start_date, end_date, daily_limit, status, created_at, updated_at
)
VALUES (
    gen_random_uuid(),
    (SELECT id FROM users WHERE email = 'aqil@gmail.com'),
    'Worker Test AI Campaign - ' || NOW()::TEXT,
    'Manual',
    'Active',
    1, 5, 100,
    CURRENT_DATE,
    CURRENT_DATE + INTERVAL '7 days',
    1000,
    'active',
    NOW(),
    NOW()
);
        """)
        
    def check_worker_endpoints(self):
        """Check if worker status endpoints exist"""
        self.log("\n" + "=" * 60)
        self.log("🔌 CHECKING WORKER STATUS ENDPOINTS")
        self.log("=" * 60)
        
        endpoints = [
            "/api/v1/worker/status",
            "/api/v1/worker/campaign/status",
            "/api/v1/worker/sequence/status",
            "/api/v1/worker/ai-campaign/status",
            "/worker/status",
            "/status"
        ]
        
        for endpoint in endpoints:
            try:
                response = self.session.get(f"{self.base_url}{endpoint}", timeout=5)
                if response.status_code == 200:
                    self.log(f"✅ Found worker endpoint: {endpoint}", "SUCCESS")
                    self.log(f"   Response: {response.text[:100]}...")
                elif response.status_code == 401:
                    self.log(f"🔒 {endpoint} requires auth", "WARNING")
                else:
                    self.log(f"❌ {endpoint} returned {response.status_code}", "DEBUG")
            except:
                pass
                
    def run_test(self):
        """Run all worker tests"""
        self.log("🚀 WORKER PROCESSING TEST FOR 3000 DEVICES")
        self.log("=" * 60)
        self.log(f"System: {self.base_url}")
        self.log(f"Time: {datetime.now()}")
        self.log("")
        
        # Check logs
        self.check_logs()
        
        # Database queries
        self.check_database_activity()
        
        # Test data SQL
        self.generate_test_data_sql()
        
        # Check endpoints
        self.check_worker_endpoints()
        
        # Summary
        self.log("\n" + "=" * 60)
        self.log("📊 HOW TO VERIFY WORKERS ARE PROCESSING:")
        self.log("=" * 60)
        
        self.log("\n1. CHECK RAILWAY LOGS:")
        self.log("   - Go to Railway dashboard")
        self.log("   - Click on your app")
        self.log("   - Go to 'Logs' tab")
        self.log("   - Look for worker messages listed above")
        
        self.log("\n2. RUN DATABASE QUERIES:")
        self.log("   - Connect to PostgreSQL")
        self.log("   - Run the queries above")
        self.log("   - Check if counts are increasing")
        
        self.log("\n3. MONITOR IN REAL-TIME:")
        self.log("   - Watch broadcast_messages table")
        self.log("   - Monitor sequence_contacts updates")
        self.log("   - Check ai_campaign_leads assignments")
        
        self.log("\n4. EXPECTED BEHAVIOR:")
        self.log("   ✓ Campaign worker runs every 30 seconds")
        self.log("   ✓ Sequence worker runs every 60 seconds")
        self.log("   ✓ AI campaign worker runs every 60 seconds")
        self.log("   ✓ Messages should appear in database tables")
        self.log("   ✓ Device last_seen should update")
        
        self.log("\n⚠️ IF WORKERS NOT RUNNING:")
        self.log("   1. Check if workers are enabled in code")
        self.log("   2. Verify database connection in logs")
        self.log("   3. Ensure devices are online")
        self.log("   4. Check for error messages in logs")

# Run the test
if __name__ == "__main__":
    tester = WorkerProcessingTest()
    tester.run_test()
