import psycopg2
from datetime import datetime

# Connect to PostgreSQL
print("Connecting to PostgreSQL...")
conn = psycopg2.connect(
    "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"
)
cursor = conn.cursor()
print("Connected successfully!")

# Task 1: Update sequence_id where it's NULL but sequence_stepid exists
print("\n=== Task 1: Fixing NULL sequence_id ===")
cursor.execute("""
    SELECT COUNT(*) 
    FROM broadcast_messages 
    WHERE sequence_id IS NULL AND sequence_stepid IS NOT NULL
""")
before_count = cursor.fetchone()[0]
print(f"Records with NULL sequence_id but NOT NULL sequence_stepid: {before_count}")

if before_count > 0:
    cursor.execute("""
        UPDATE broadcast_messages bm
        SET sequence_id = ss.sequence_id
        FROM sequence_steps ss
        WHERE bm.sequence_stepid = ss.id
        AND bm.sequence_id IS NULL 
        AND bm.sequence_stepid IS NOT NULL
    """)
    updated_count = cursor.rowcount
    conn.commit()
    print(f"Updated {updated_count} records with missing sequence_id")
else:
    print("No records to update for Task 1")

# Task 2: Fix failed messages with "no campaign ID or sequence step ID" error
print("\n=== Task 2: Fixing failed messages ===")
cursor.execute("""
    SELECT COUNT(*) 
    FROM broadcast_messages 
    WHERE status = 'failed' 
    AND error_message = 'message has no campaign ID or sequence step ID'
""")
failed_count = cursor.fetchone()[0]
print(f"Failed messages with 'no campaign ID or sequence step ID' error: {failed_count}")

if failed_count > 0:
    cursor.execute("""
        UPDATE broadcast_messages 
        SET status = 'pending', error_message = NULL
        WHERE status = 'failed' 
        AND error_message = 'message has no campaign ID or sequence step ID'
    """)
    updated_failed = cursor.rowcount
    conn.commit()
    print(f"Reset {updated_failed} failed messages to pending status")
else:
    print("No failed messages to update")

# Task 3: Analyze timeout messages
print("\n=== Task 3: Analyzing timeout messages ===")
cursor.execute("""
    SELECT COUNT(*) 
    FROM broadcast_messages 
    WHERE status = 'sent' 
    AND error_message = 'Message timeout - device was not available'
""")
timeout_count = cursor.fetchone()[0]
print(f"Messages marked as 'sent' with timeout error: {timeout_count}")

# Check how many of these have platform devices
cursor.execute("""
    SELECT COUNT(DISTINCT bm.device_id), 
           COUNT(CASE WHEN ud.platform IS NOT NULL AND ud.platform != '' THEN 1 END) as platform_count
    FROM broadcast_messages bm
    LEFT JOIN user_devices ud ON bm.device_id = ud.id
    WHERE bm.status = 'sent' 
    AND bm.error_message = 'Message timeout - device was not available'
""")
device_info = cursor.fetchone()
print(f"Unique devices with timeout: {device_info[0]}")
print(f"Platform devices with timeout: {device_info[1]}")

# Get sample of platform devices with timeout
cursor.execute("""
    SELECT DISTINCT ud.phone, ud.platform, COUNT(bm.id) as message_count
    FROM broadcast_messages bm
    JOIN user_devices ud ON bm.device_id = ud.id
    WHERE bm.status = 'sent' 
    AND bm.error_message = 'Message timeout - device was not available'
    AND ud.platform IS NOT NULL AND ud.platform != ''
    GROUP BY ud.phone, ud.platform
    LIMIT 10
""")
platform_timeouts = cursor.fetchall()

if platform_timeouts:
    print("\nSample platform devices with timeout errors:")
    for row in platform_timeouts:
        print(f"  Phone: {row[0]}, Platform: {row[1]}, Messages: {row[2]}")

# Close connection
cursor.close()
conn.close()
print("\n✅ All tasks completed!")
