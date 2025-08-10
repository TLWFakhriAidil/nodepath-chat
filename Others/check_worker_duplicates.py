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

print("Investigating why device worker is sending duplicates...")
print("="*100)

# First, check if the processing_worker_id column exists and is being used
cursor.execute("SHOW COLUMNS FROM broadcast_messages LIKE 'processing_worker_id'")
column_exists = cursor.fetchone()
print(f"\nprocessing_worker_id column exists: {column_exists is not None}")

if column_exists:
    # Check if it's actually being populated
    cursor.execute("""
    SELECT 
        COUNT(*) as total,
        COUNT(processing_worker_id) as with_worker_id,
        COUNT(CASE WHEN processing_worker_id IS NULL THEN 1 END) as without_worker_id
    FROM broadcast_messages 
    WHERE created_at >= DATE_SUB(NOW(), INTERVAL 1 DAY)
    """)
    
    stats = cursor.fetchone()
    print(f"\nLast 24 hours statistics:")
    print(f"  Total messages: {stats[0]}")
    print(f"  With worker ID: {stats[1]}")
    print(f"  Without worker ID (NULL): {stats[2]}")

# Check for the same message being sent multiple times
print("\n" + "="*100)
print("Checking for same message sent multiple times...")

# Look for messages with same content sent to same recipient around same time
cursor.execute("""
SELECT 
    recipient_phone,
    LEFT(content, 50) as content_preview,
    COUNT(*) as send_count,
    MIN(sent_at) as first_sent,
    MAX(sent_at) as last_sent,
    TIMESTAMPDIFF(SECOND, MIN(sent_at), MAX(sent_at)) as seconds_apart,
    GROUP_CONCAT(id ORDER BY sent_at) as message_ids
FROM broadcast_messages 
WHERE sent_at >= DATE_SUB(NOW(), INTERVAL 2 DAY)
AND status = 'sent'
GROUP BY recipient_phone, LEFT(content, 50)
HAVING COUNT(*) > 1
AND TIMESTAMPDIFF(SECOND, MIN(sent_at), MAX(sent_at)) < 300  -- Within 5 minutes
ORDER BY COUNT(*) DESC, MAX(sent_at) DESC
LIMIT 10
""")

duplicates = cursor.fetchall()
print(f"\nFound {len(duplicates)} cases of duplicate sending:")

for dup in duplicates:
    print(f"\n  Phone: {dup[0]}")
    print(f"  Content: {dup[1]}...")
    print(f"  Sent {dup[2]} times")
    print(f"  First: {dup[3]}, Last: {dup[4]} ({dup[5]} seconds apart)")
    
    # Get details of each message
    ids = dup[6].split(',')
    for msg_id in ids[:3]:  # First 3
        cursor.execute("""
        SELECT status, processing_worker_id, device_id, created_at, sent_at
        FROM broadcast_messages WHERE id = %s
        """, (msg_id.strip(),))
        details = cursor.fetchone()
        if details:
            print(f"    - ID {msg_id[:8]}...: status={details[0]}, worker={details[1]}, created={details[3]}")

conn.close()
