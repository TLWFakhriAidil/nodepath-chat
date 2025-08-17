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

print("Analyzing duplicate messages from screenshot (Aug 7, 6:10 AM)...")
print("="*100)

# Based on the screenshot, messages were sent at 6:10 AM on Aug 7
# Let's check for messages around that time
query = """
SELECT 
    id,
    recipient_phone,
    sequence_id,
    sequence_stepid,
    status,
    created_at,
    scheduled_at,
    sent_at,
    processing_worker_id,
    device_id,
    LEFT(content, 80) as content_preview
FROM broadcast_messages 
WHERE recipient_phone = '601139938358'
AND DATE(sent_at) >= '2025-08-06'
ORDER BY sent_at DESC
"""

cursor.execute(query)
results = cursor.fetchall()

print(f"Found {len(results)} messages for 601139938358")
print("\nDetailed analysis:")

# Group by time to find duplicates
time_groups = {}
for r in results:
    if r[7]:  # if sent_at exists
        # Group by minute (ignoring seconds)
        time_key = r[7].strftime("%Y-%m-%d %H:%M")
        if time_key not in time_groups:
            time_groups[time_key] = []
        time_groups[time_key].append(r)

# Show groups with multiple messages
for time_key, messages in time_groups.items():
    if len(messages) > 1:
        print(f"\n{'='*80}")
        print(f"DUPLICATE FOUND at {time_key} - {len(messages)} messages:")
        
        for i, msg in enumerate(messages):
            print(f"\n  Message {i+1}:")
            print(f"    ID: {msg[0]}")
            print(f"    Step ID: {msg[3][:20] if msg[3] else 'None'}...")
            print(f"    Created: {msg[5]}")
            print(f"    Scheduled: {msg[6]}")
            print(f"    Sent: {msg[7]}")
            print(f"    Worker: {msg[8]}")
            print(f"    Device: {msg[9][:8] if msg[9] else 'None'}...")
            print(f"    Content: {msg[10][:50].encode('ascii', 'ignore').decode()}...")

# Check worker ID implementation
print("\n" + "="*100)
print("Checking worker ID implementation for today's messages...")

cursor.execute("""
SELECT 
    COUNT(*) as total_messages,
    COUNT(DISTINCT sequence_stepid) as unique_steps,
    COUNT(processing_worker_id) as messages_with_worker,
    COUNT(DISTINCT processing_worker_id) as unique_workers
FROM broadcast_messages 
WHERE sequence_stepid IS NOT NULL
AND DATE(created_at) = CURDATE()
""")

stats = cursor.fetchone()
print(f"\nToday's sequence messages:")
print(f"  Total messages: {stats[0]}")
print(f"  Unique steps: {stats[1]}")
print(f"  Messages with worker ID: {stats[2]}")
print(f"  Unique worker IDs: {stats[3]}")

# Check if duplicate prevention is working at creation time
print("\n" + "="*100)
print("Checking duplicate prevention at message creation...")

cursor.execute("""
SELECT 
    sequence_stepid,
    recipient_phone,
    COUNT(*) as count,
    MIN(created_at) as first_created,
    MAX(created_at) as last_created,
    TIMESTAMPDIFF(SECOND, MIN(created_at), MAX(created_at)) as seconds_apart
FROM broadcast_messages 
WHERE sequence_stepid IS NOT NULL
AND DATE(created_at) >= DATE_SUB(CURDATE(), INTERVAL 2 DAY)
GROUP BY sequence_stepid, recipient_phone
HAVING COUNT(*) > 1
ORDER BY COUNT(*) DESC
LIMIT 10
""")

duplicates = cursor.fetchall()
print(f"\nFound {cursor.rowcount} duplicate step/phone combinations in last 2 days:")

for dup in duplicates:
    print(f"\n  Step: {dup[0][:20]}...")
    print(f"  Phone: {dup[1]}")
    print(f"  Count: {dup[2]}")
    print(f"  First created: {dup[3]}")
    print(f"  Last created: {dup[4]}")
    print(f"  Time apart: {dup[5]} seconds")

conn.close()
