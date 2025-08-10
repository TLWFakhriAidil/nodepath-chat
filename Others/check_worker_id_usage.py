import pymysql
from datetime import datetime, timedelta

conn = pymysql.connect(
    host='159.89.198.71',
    user='admin_aqil',
    password='admin_aqil',
    database='admin_railway',
    port=3306
)
cursor = conn.cursor()

print("CHECKING: Are messages for Aug 7-8 using GetPendingMessagesAndLock?")
print("="*80)

# Check messages scheduled for Aug 7-8
cursor.execute("""
SELECT 
    DATE(scheduled_at) as scheduled_date,
    COUNT(*) as total_messages,
    COUNT(processing_worker_id) as with_worker_id,
    COUNT(processing_started_at) as with_start_time,
    MIN(created_at) as earliest_created,
    MAX(created_at) as latest_created
FROM broadcast_messages 
WHERE DATE(scheduled_at) IN ('2025-08-07', '2025-08-08')
GROUP BY DATE(scheduled_at)
ORDER BY scheduled_date
""")

results = cursor.fetchall()
print("\nMessages scheduled for Aug 7-8:")
for r in results:
    print(f"\nDate: {r[0]}")
    print(f"  Total messages: {r[1]}")
    print(f"  With worker ID: {r[2]} ({r[2]/r[1]*100:.1f}%)")
    print(f"  With start time: {r[3]} ({r[3]/r[1]*100:.1f}%)")
    print(f"  Created between: {r[4]} to {r[5]}")

# Check specific pending messages
print("\n" + "="*80)
print("Checking PENDING messages for Aug 7-8:")

cursor.execute("""
SELECT 
    id,
    status,
    processing_worker_id,
    processing_started_at,
    scheduled_at,
    created_at,
    device_id,
    sequence_stepid,
    campaign_id
FROM broadcast_messages 
WHERE DATE(scheduled_at) IN ('2025-08-07', '2025-08-08')
AND status = 'pending'
LIMIT 10
""")

pending = cursor.fetchall()
print(f"\nFound {cursor.rowcount} pending messages (showing first 10):")

for msg in pending:
    print(f"\n  ID: {msg[0][:8]}...")
    print(f"    Status: {msg[1]}")
    print(f"    Worker ID: {msg[2] or 'NULL'}")
    print(f"    Started at: {msg[3] or 'NULL'}")
    print(f"    Scheduled: {msg[4]}")
    print(f"    Type: {'Sequence' if msg[7] else 'Campaign' if msg[8] else 'Unknown'}")

# Check if GetPendingMessagesAndLock is working
print("\n" + "="*80)
print("VERIFICATION: Is GetPendingMessagesAndLock being called?")

cursor.execute("""
SELECT 
    COUNT(*) as total_all_time,
    COUNT(processing_worker_id) as with_worker_all_time
FROM broadcast_messages
""")

all_time = cursor.fetchone()
print(f"\nAll-time statistics:")
print(f"  Total messages: {all_time[0]}")
print(f"  With worker ID: {all_time[1]} ({all_time[1]/all_time[0]*100:.1f}%)")

if all_time[1] == 0:
    print("\n*** PROBLEM: No messages have worker IDs! ***")
    print("GetPendingMessagesAndLock is NOT being called!")
    print("\nPossible reasons:")
    print("1. The fix hasn't been deployed yet")
    print("2. The application needs to be restarted")
    print("3. There's another place calling GetPendingMessages")

conn.close()
