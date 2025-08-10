import pymysql

# Database connection
connection = pymysql.connect(
    host='159.89.198.71',
    user='admin_aqil',
    password='admin_aqil',
    database='admin_railway',
    charset='utf8mb4',
    cursorclass=pymysql.cursors.DictCursor
)

try:
    with connection.cursor() as cursor:
        print("=== CHECKING WHY MESSAGES WERE CREATED TWICE ===\n")
        
        # Check the two specific messages
        query = """
        SELECT 
            id,
            recipient_phone,
            device_id,
            sequence_id,
            sequence_stepid,
            created_at,
            scheduled_at,
            sent_at,
            status
        FROM broadcast_messages
        WHERE id IN ('e8ef71e1-2def-41af-ba1f-ccc3111b5d21', 'a23a9647-931b-46bf-a692-d5c5b14f06e7')
        """
        
        cursor.execute(query)
        messages = cursor.fetchall()
        
        for msg in messages:
            print(f"Message ID: {msg['id']}")
            print(f"  Phone: {msg['recipient_phone']}")
            print(f"  Device: {msg['device_id']}")
            print(f"  Sequence: {msg['sequence_id']}")
            print(f"  Step ID: {msg['sequence_stepid']}")
            print(f"  Created: {msg['created_at']}")
            print(f"  Scheduled: {msg['scheduled_at']}")
            print(f"  Sent: {msg['sent_at']}")
            print(f"  Status: {msg['status']}")
            print()
        
        # Check if there are more duplicates at creation time
        print("\n=== CHECKING ENROLLMENT PATTERNS ===\n")
        
        query = """
        SELECT 
            recipient_phone,
            sequence_stepid,
            COUNT(*) as msg_count,
            COUNT(DISTINCT device_id) as device_count,
            MIN(created_at) as first_created,
            MAX(created_at) as last_created,
            TIMESTAMPDIFF(SECOND, MIN(created_at), MAX(created_at)) as seconds_apart
        FROM broadcast_messages
        WHERE sequence_stepid = '72e96e33-2169-4d72-8d2d-041eab647e53'
        GROUP BY recipient_phone, sequence_stepid
        HAVING COUNT(*) > 1
        ORDER BY msg_count DESC
        LIMIT 10
        """
        
        cursor.execute(query)
        duplicates = cursor.fetchall()
        
        print(f"Recipients enrolled multiple times in same step:\n")
        
        for dup in duplicates:
            print(f"Phone: {dup['recipient_phone']}")
            print(f"  Messages created: {dup['msg_count']}")
            print(f"  Devices assigned: {dup['device_count']}")
            print(f"  Created: {dup['first_created']} to {dup['last_created']}")
            print(f"  Time apart: {dup['seconds_apart']} seconds")
            print()

finally:
    connection.close()
