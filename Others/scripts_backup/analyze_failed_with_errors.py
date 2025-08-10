import psycopg2
import time

# Connect with retry
conn_str = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

for attempt in range(3):
    try:
        print(f"Connection attempt {attempt + 1}...")
        conn = psycopg2.connect(conn_str, connect_timeout=30)
        cursor = conn.cursor()
        print("[OK] Connected!")
        break
    except Exception as e:
        print(f"Failed: {e}")
        if attempt < 2:
            time.sleep(5)

# First, let's check the exact columns in broadcast_messages
print("\nChecking broadcast_messages table structure...")
cursor.execute("""
    SELECT column_name, data_type 
    FROM information_schema.columns 
    WHERE table_name = 'broadcast_messages'
    ORDER BY ordinal_position;
""")

columns = cursor.fetchall()
print("\nColumns in broadcast_messages:")
for col, dtype in columns:
    print(f"  - {col}: {dtype}")

# Now query with correct columns
print("\n" + "="*80)
print("FAILED SEQUENCE MESSAGES WITH ERRORS")
print("="*80)

# Updated query with correct column names
query = """
SELECT 
    bm.id,
    bm.recipient_phone,
    LEFT(bm.message, 80) as message_preview,
    bm.error_message,
    bm.status,
    bm.created_at,
    bm.updated_at,
    bm.device_id,
    bm.sequence_stepid,
    ud.device_name,
    ud.phone as device_phone,
    s.name as sequence_name
FROM broadcast_messages bm
LEFT JOIN user_devices ud ON bm.device_id = ud.id
LEFT JOIN sequence_steps ss ON bm.sequence_stepid = ss.id
LEFT JOIN sequences s ON ss.sequence_id = s.id
WHERE bm.sequence_stepid IS NOT NULL 
AND bm.status = 'failed' 
AND bm.error_message IS NOT NULL
ORDER BY bm.updated_at DESC
LIMIT 15;
"""

cursor.execute(query)
results = cursor.fetchall()

print(f"\nFound {len(results)} failed sequence messages with errors")

if results:
    for i, row in enumerate(results, 1):
        print(f"\n[Message {i}]")
        print(f"  ID: {row[0]}")
        print(f"  Recipient: {row[1]}")
        print(f"  Message: {row[2]}...")
        print(f"  ERROR: {row[3]}")
        print(f"  Created: {row[5]}")
        print(f"  Failed at: {row[6]}")
        print(f"  Device: {row[9] or 'Unknown'} ({row[10] or 'No phone'})")
        print(f"  Sequence: {row[11] or 'Unknown'}")

# Get error distribution
print("\n" + "="*80)
print("ERROR MESSAGE DISTRIBUTION")
print("="*80)

cursor.execute("""
    SELECT error_message, COUNT(*) as count
    FROM broadcast_messages 
    WHERE sequence_stepid IS NOT NULL 
    AND status = 'failed' 
    AND error_message IS NOT NULL
    GROUP BY error_message
    ORDER BY count DESC
""")

errors = cursor.fetchall()
total_errors = sum(e[1] for e in errors)

print(f"\nTotal failed sequence messages: {total_errors}")
print("\nError breakdown:")
for error, count in errors:
    percentage = (count / total_errors) * 100
    print(f"\n  [{count} messages - {percentage:.1f}%]")
    print(f"  {error}")

# Check device status for failed messages
print("\n" + "="*80)
print("DEVICE STATUS FOR FAILED MESSAGES")
print("="*80)

cursor.execute("""
    SELECT 
        COALESCE(ud.device_name, 'No Device') as device,
        ud.status as device_status,
        COUNT(bm.id) as failed_count
    FROM broadcast_messages bm
    LEFT JOIN user_devices ud ON bm.device_id = ud.id
    WHERE bm.sequence_stepid IS NOT NULL 
    AND bm.status = 'failed' 
    AND bm.error_message IS NOT NULL
    GROUP BY ud.device_name, ud.status
    ORDER BY failed_count DESC
""")

device_stats = cursor.fetchall()
for device, status, count in device_stats:
    print(f"  {device} [{status or 'Unknown'}]: {count} failures")

cursor.close()
conn.close()
print("\n[OK] Analysis complete!")
