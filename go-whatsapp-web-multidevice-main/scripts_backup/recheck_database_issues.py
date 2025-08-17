import psycopg2
import sys
from datetime import datetime

# Set UTF-8 encoding for output
sys.stdout.reconfigure(encoding='utf-8')

# Connect to PostgreSQL
print("Connecting to PostgreSQL...")
conn = psycopg2.connect(
    "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"
)
cursor = conn.cursor()
print("Connected successfully!")
print(f"Check time: {datetime.now()}")

print("\n" + "="*60)
print("RECHECKING DATABASE FOR ISSUES")
print("="*60)

# Issue 1: Check for NULL sequence_id but NOT NULL sequence_stepid
print("\n1. Checking for NULL sequence_id with NOT NULL sequence_stepid:")
cursor.execute("""
    SELECT COUNT(*) 
    FROM broadcast_messages 
    WHERE sequence_id IS NULL AND sequence_stepid IS NOT NULL
""")
count1 = cursor.fetchone()[0]
print(f"   Found: {count1} records")

if count1 > 0:
    print("   ⚠️  New records found with this issue!")
    cursor.execute("""
        SELECT id, created_at, status, recipient_phone
        FROM broadcast_messages 
        WHERE sequence_id IS NULL AND sequence_stepid IS NOT NULL
        LIMIT 5
    """)
    samples = cursor.fetchall()
    for sample in samples:
        print(f"      - ID: {sample[0]}, Created: {sample[1]}, Status: {sample[2]}, Phone: {sample[3]}")

# Issue 2: Check for failed messages with specific error
print("\n2. Checking for failed messages with 'no campaign ID or sequence step ID' error:")
cursor.execute("""
    SELECT COUNT(*) 
    FROM broadcast_messages 
    WHERE status = 'failed' 
    AND error_message = 'message has no campaign ID or sequence step ID'
""")
count2 = cursor.fetchone()[0]
print(f"   Found: {count2} records")

if count2 > 0:
    print("   ⚠️  New failed messages found!")
    cursor.execute("""
        SELECT id, created_at, recipient_phone
        FROM broadcast_messages 
        WHERE status = 'failed' 
        AND error_message = 'message has no campaign ID or sequence step ID'
        LIMIT 5
    """)
    samples = cursor.fetchall()
    for sample in samples:
        print(f"      - ID: {sample[0]}, Created: {sample[1]}, Phone: {sample[2]}")

# Issue 3: Check for platform device timeout errors
print("\n3. Checking for platform devices with timeout errors:")
cursor.execute("""
    SELECT COUNT(bm.id), COUNT(DISTINCT ud.id), STRING_AGG(DISTINCT ud.platform, ', ') as platforms
    FROM broadcast_messages bm
    JOIN user_devices ud ON bm.device_id = ud.id
    WHERE bm.status = 'sent' 
    AND bm.error_message = 'Message timeout - device was not available'
    AND ud.platform IS NOT NULL AND ud.platform != ''
""")
result = cursor.fetchone()
count3 = result[0] if result[0] else 0
device_count = result[1] if result[1] else 0
platforms = result[2] if result[2] else 'None'

print(f"   Found: {count3} messages from {device_count} platform devices")
if count3 > 0:
    print(f"   Platforms affected: {platforms}")
    print("   ⚠️  Platform timeout issue still exists!")
    
    # Get recent examples
    cursor.execute("""
        SELECT bm.id, bm.created_at, ud.platform, bm.recipient_phone
        FROM broadcast_messages bm
        JOIN user_devices ud ON bm.device_id = ud.id
        WHERE bm.status = 'sent' 
        AND bm.error_message = 'Message timeout - device was not available'
        AND ud.platform IS NOT NULL AND ud.platform != ''
        ORDER BY bm.created_at DESC
        LIMIT 5
    """)
    recent = cursor.fetchall()
    if recent:
        print("   Recent examples:")
        for r in recent:
            print(f"      - ID: {r[0]}, Created: {r[1]}, Platform: {r[2]}, Phone: {r[3]}")

# Check for any WhatsApp Web devices with timeout (this is expected behavior)
print("\n4. Checking WhatsApp Web devices with timeout (expected behavior):")
cursor.execute("""
    SELECT COUNT(bm.id), COUNT(DISTINCT ud.id)
    FROM broadcast_messages bm
    JOIN user_devices ud ON bm.device_id = ud.id
    WHERE bm.status = 'sent' 
    AND bm.error_message = 'Message timeout - device was not available'
    AND (ud.platform IS NULL OR ud.platform = '')
""")
result = cursor.fetchone()
wa_timeout = result[0] if result[0] else 0
wa_devices = result[1] if result[1] else 0
print(f"   Found: {wa_timeout} messages from {wa_devices} WhatsApp Web devices")
print("   ✅ This is expected - WhatsApp Web devices can timeout when offline")

# Summary
print("\n" + "="*60)
print("SUMMARY")
print("="*60)

total_issues = count1 + count2 + count3
if total_issues == 0:
    print("✅ ALL ISSUES RESOLVED! No problems found.")
else:
    print(f"⚠️  ISSUES FOUND: {total_issues} total records need attention")
    print("\nRecommendations:")
    if count1 > 0:
        print(f"- Fix {count1} records with NULL sequence_id")
    if count2 > 0:
        print(f"- Reset {count2} failed messages to pending")
    if count3 > 0:
        print(f"- Platform timeout issue needs investigation - {count3} messages affected")
        print("  The code fix was applied but new timeouts may have occurred")

# Close connection
cursor.close()
conn.close()
print("\nDatabase check completed!")
