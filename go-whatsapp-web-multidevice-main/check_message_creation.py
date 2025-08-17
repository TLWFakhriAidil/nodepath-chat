import pymysql
from datetime import datetime

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
        print("=== CHECKING FOR DUPLICATE MESSAGE IDs ===\n")
        
        # Check if there are duplicate message IDs that were both sent
        query = """
        SELECT 
            id,
            COUNT(*) as count,
            GROUP_CONCAT(status) as statuses,
            GROUP_CONCAT(sent_at) as sent_times
        FROM broadcast_messages
        WHERE DATE(sent_at) = '2025-08-06'
        GROUP BY id
        HAVING COUNT(*) > 1
        """
        
        cursor.execute(query)
        duplicates = cursor.fetchall()
        
        if duplicates:
            print(f"CRITICAL: Found {len(duplicates)} duplicate message IDs!")
            for dup in duplicates:
                print(f"  ID: {dup['id']}")
                print(f"  Count: {dup['count']}")
                print(f"  Statuses: {dup['statuses']}")
                print(f"  Sent times: {dup['sent_times']}")
        else:
            print("No duplicate message IDs found")
        
        # Check for messages queued multiple times
        print("\n\n=== CHECKING MESSAGE STATUS HISTORY ===\n")
        
        # Look at some example messages that were sent on Aug 6
        query = """
        SELECT 
            id,
            recipient_phone,
            status,
            created_at,
            scheduled_at,
            sent_at,
            updated_at,
            sequence_stepid
        FROM broadcast_messages
        WHERE recipient_phone = '60122712014'
            AND DATE(sent_at) = '2025-08-06'
        ORDER BY created_at
        LIMIT 10
        """
        
        cursor.execute(query)
        messages = cursor.fetchall()
        
        for msg in messages:
            print(f"Message: {msg['id'][:8]}...")
            print(f"  Phone: {msg['recipient_phone']}")
            print(f"  Status: {msg['status']}")
            print(f"  Created: {msg['created_at']}")
            print(f"  Updated: {msg['updated_at']}")
            print(f"  Sent: {msg['sent_at']}")
            print(f"  Step ID: {msg['sequence_stepid']}")
            print()
        
        # Check if messages are being created multiple times
        print("\n=== CHECKING FOR DUPLICATE CREATION ===\n")
        
        query = """
        SELECT 
            recipient_phone,
            sequence_stepid,
            COUNT(*) as msg_count,
            GROUP_CONCAT(id) as message_ids,
            GROUP_CONCAT(status) as statuses,
            MIN(created_at) as first_created,
            MAX(created_at) as last_created
        FROM broadcast_messages
        WHERE sequence_stepid IS NOT NULL
            AND DATE(sent_at) = '2025-08-06'
        GROUP BY recipient_phone, sequence_stepid
        HAVING COUNT(*) > 1
        ORDER BY msg_count DESC
        LIMIT 10
        """
        
        cursor.execute(query)
        dup_creations = cursor.fetchall()
        
        print(f"Found {len(dup_creations)} cases of duplicate message creation\n")
        
        for dup in dup_creations:
            print(f"Phone: {dup['recipient_phone']}")
            print(f"  Same Step ID: {dup['sequence_stepid']}")
            print(f"  Messages created: {dup['msg_count']}")
            print(f"  IDs: {dup['message_ids'][:50]}...")
            print(f"  Statuses: {dup['statuses']}")
            print(f"  Time span: {dup['first_created']} to {dup['last_created']}")
            print()

finally:
    connection.close()
