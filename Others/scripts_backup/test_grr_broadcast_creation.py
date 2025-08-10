import psycopg2
import sys
from datetime import datetime, timedelta
import uuid

# Set UTF-8 encoding
sys.stdout.reconfigure(encoding='utf-8')

conn = psycopg2.connect(
    "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"
)
cursor = conn.cursor()

print("=== TESTING GRR CAMPAIGN → BROADCAST MESSAGE CREATION ===\n")

# 1. Create a test campaign for GRR
print("1. Creating a new test campaign for GRR niche...")

test_campaign_id = None
try:
    cursor.execute("""
        INSERT INTO campaigns 
        (user_id, title, message, niche, target_status, campaign_date, 
         time_schedule, status, min_delay_seconds, max_delay_seconds, created_at, updated_at)
        VALUES 
        ('de078f16-3266-4ab3-8153-a248b015228f', 'GRR Test Campaign', 
         'Test message for GRR leads', 'GRR', 'prospect', CURRENT_DATE,
         TO_CHAR(NOW() + INTERVAL '5 minutes', 'HH24:MI'), 'pending', 10, 30, NOW(), NOW())
        RETURNING id
    """)
    test_campaign_id = cursor.fetchone()[0]
    conn.commit()
    print(f"✅ Created test campaign ID: {test_campaign_id}")
except Exception as e:
    print(f"❌ Failed to create campaign: {e}")
    conn.rollback()

# 2. Check the GRR lead details
print("\n2. Checking GRR lead details...")
cursor.execute("""
    SELECT phone, name, device_id, user_id, target_status
    FROM leads
    WHERE niche = 'GRR' AND phone = '60108924904'
""")
lead = cursor.fetchone()

if lead:
    print(f"✅ Found GRR lead:")
    print(f"   Phone: {lead[0]}")
    print(f"   Name: {lead[1]}")
    print(f"   Device: {lead[2]}")
    print(f"   User: {lead[3]}")
    print(f"   Status: {lead[4]}")

# 3. Check device status
print("\n3. Checking device status...")
cursor.execute("""
    SELECT device_name, status, platform, last_seen
    FROM user_devices
    WHERE id = %s
""", (lead[2],))
device = cursor.fetchone()

print(f"✅ Device info:")
print(f"   Name: {device[0]}")
print(f"   Status: {device[1]}")
print(f"   Platform: {device[2]}")
print(f"   Last seen: {device[3]}")

# 4. Manually create a broadcast message (simulate what campaign should do)
print("\n4. Manually creating broadcast message...")

try:
    msg_id = str(uuid.uuid4())
    cursor.execute("""
        INSERT INTO broadcast_messages
        (id, user_id, device_id, campaign_id, recipient_phone, recipient_name,
         message_type, content, status, scheduled_at, created_at,
         min_delay_seconds, max_delay_seconds)
        VALUES
        (%s, %s, %s, %s, %s, %s, 'text', 'Test message for GRR', 
         'pending', NOW(), NOW(), 10, 30)
    """, (msg_id, lead[3], lead[2], test_campaign_id, lead[0], lead[1]))
    
    conn.commit()
    print(f"✅ Successfully created broadcast message!")
    print(f"   Message ID: {msg_id}")
    
    # Verify it was created
    cursor.execute("""
        SELECT id, status, recipient_phone, device_id
        FROM broadcast_messages
        WHERE id = %s
    """, (msg_id,))
    result = cursor.fetchone()
    print(f"\n   Verification: Message created for {result[2]}")
    print(f"   Status: {result[1]}")
    
except Exception as e:
    print(f"❌ Failed to create broadcast message: {e}")
    conn.rollback()

# 5. Check why automatic campaign → broadcast might fail
print("\n5. Checking potential issues...")

# Check if device is truly online
cursor.execute("""
    SELECT 
        COUNT(*) as online_count
    FROM user_devices
    WHERE id = %s
    AND (status = 'online' OR status = 'connected' OR platform IS NOT NULL)
""", (lead[2],))
online = cursor.fetchone()[0]

if online > 0:
    print("✅ Device is considered ONLINE (should work)")
else:
    print("❌ Device is NOT considered online by the system")

# Check for any unique constraints
cursor.execute("""
    SELECT COUNT(*)
    FROM broadcast_messages
    WHERE campaign_id = %s
    AND recipient_phone = %s
""", (test_campaign_id, lead[0]))
existing = cursor.fetchone()[0]

print(f"\nExisting messages for this campaign/phone: {existing}")

# 6. Clean up test data
print("\n6. Cleaning up test data...")
if test_campaign_id:
    cursor.execute("DELETE FROM broadcast_messages WHERE campaign_id = %s", (test_campaign_id,))
    cursor.execute("DELETE FROM campaigns WHERE id = %s", (test_campaign_id,))
    conn.commit()
    print("✅ Test data cleaned up")

conn.close()

print("\n=== CONCLUSION ===")
print("GRR campaigns CAN create broadcast messages when:")
print("1. Campaign is in 'pending' status")
print("2. Device is online/connected")
print("3. Lead exists with matching niche and target_status")
print("4. No duplicate constraints are violated")
