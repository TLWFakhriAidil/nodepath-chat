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

print("Checking for duplicate sequence messages...")
print("="*100)

# Check messages sent today with duplicates
query = """
SELECT 
    recipient_phone,
    sequence_stepid,
    COUNT(*) as count,
    GROUP_CONCAT(id) as message_ids,
    GROUP_CONCAT(status) as statuses,
    GROUP_CONCAT(sent_at) as sent_times,
    GROUP_CONCAT(device_id) as devices,
    GROUP_CONCAT(processing_worker_id) as workers
FROM broadcast_messages 
WHERE sequence_stepid IS NOT NULL
AND DATE(sent_at) = CURDATE()
GROUP BY recipient_phone, sequence_stepid
HAVING COUNT(*) > 1
ORDER BY COUNT(*) DESC
LIMIT 20
"""

cursor.execute(query)
duplicates = cursor.fetchall()

print(f"\nFound {len(duplicates)} duplicate sequence messages sent today")

for dup in duplicates:
    phone, step_id, count, msg_ids, statuses, sent_times, devices, workers = dup
    print(f"\n{'='*80}")
    print(f"Phone: {phone}")
    print(f"Step ID: {step_id[:20]}...")
    print(f"Duplicate Count: {count}")
    
    # Get more details for each duplicate
    ids = msg_ids.split(',')
    for i, msg_id in enumerate(ids[:3]):  # Show first 3
        cursor.execute("""
        SELECT created_at, scheduled_at, sent_at, status, 
               processing_worker_id, device_id, 
               LEFT(content, 100) as content
        FROM broadcast_messages 
        WHERE id = %s
        """, (msg_id.strip(),))
        
        details = cursor.fetchone()
        if details:
            print(f"\n  Message {i+1} (ID: {msg_id.strip()[:8]}...):")
            print(f"    Created: {details[0]}")
            print(f"    Scheduled: {details[1]}")
            print(f"    Sent: {details[2]}")
            print(f"    Status: {details[3]}")
            print(f"    Worker: {details[4]}")
            print(f"    Device: {details[5][:8]}...")
            print(f"    Content: {details[6][:60]}...")

# Check the specific phone number from the screenshot
print("\n" + "="*100)
print("Checking specific phone from screenshot...")

phone_variations = ['601139938358', '60 11-3993 8358', '+601139938358']
for phone in phone_variations:
    cursor.execute("""
    SELECT id, sequence_stepid, status, sent_at, device_id, 
           processing_worker_id, LEFT(content, 100) as content
    FROM broadcast_messages 
    WHERE recipient_phone = %s 
    AND DATE(sent_at) >= DATE_SUB(CURDATE(), INTERVAL 2 DAY)
    ORDER BY sent_at DESC
    LIMIT 10
    """, (phone,))
    
    results = cursor.fetchall()
    if results:
        print(f"\nFound {len(results)} messages for phone: {phone}")
        for r in results:
            print(f"\n  ID: {r[0]}")
            print(f"  Step: {r[1][:20] if r[1] else 'None'}...")
            print(f"  Status: {r[2]}")
            print(f"  Sent: {r[3]}")
            print(f"  Device: {r[4][:8] if r[4] else 'None'}...")
            print(f"  Worker: {r[5]}")
            print(f"  Content: {r[6][:60]}...")

# Check duplicate prevention mechanism
print("\n" + "="*100)
print("Checking duplicate prevention columns...")
cursor.execute("""
SELECT 
    COUNT(*) as total,
    COUNT(DISTINCT processing_worker_id) as unique_workers,
    COUNT(CASE WHEN processing_worker_id IS NULL THEN 1 END) as null_workers
FROM broadcast_messages 
WHERE sequence_stepid IS NOT NULL
AND DATE(created_at) = CURDATE()
""")

stats = cursor.fetchone()
print(f"Total sequence messages today: {stats[0]}")
print(f"Unique worker IDs: {stats[1]}")
print(f"Messages with NULL worker ID: {stats[2]}")

conn.close()
