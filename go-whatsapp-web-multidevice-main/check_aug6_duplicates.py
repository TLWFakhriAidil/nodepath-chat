import pymysql
from datetime import datetime, timedelta
from collections import defaultdict

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
        # Check all messages for August 6, 2025
        print("=== Checking for duplicate messages on August 6, 2025 ===\n")
        
        # Get all messages sent on Aug 6
        query = """
        SELECT 
            id,
            recipient_phone,
            recipient_name,
            content,
            status,
            created_at,
            scheduled_at,
            sent_at,
            device_id,
            campaign_id,
            sequence_id,
            sequence_stepid
        FROM broadcast_messages
        WHERE DATE(created_at) = '2025-08-06'
           OR DATE(scheduled_at) = '2025-08-06'
           OR DATE(sent_at) = '2025-08-06'
        ORDER BY recipient_phone, created_at
        """
        
        cursor.execute(query)
        all_messages = cursor.fetchall()
        
        print(f"Total messages found for Aug 6: {len(all_messages)}\n")
        
        # Group by recipient to find duplicates
        messages_by_recipient = defaultdict(list)
        for msg in all_messages:
            messages_by_recipient[msg['recipient_phone']].append(msg)
        
        # Find recipients with multiple messages
        duplicate_count = 0
        for phone, messages in messages_by_recipient.items():
            if len(messages) > 1:
                # Check if messages have same content
                content_groups = defaultdict(list)
                for msg in messages:
                    # Normalize content for comparison
                    content_key = (msg['content'] or '').strip()[:100]  # First 100 chars
                    content_groups[content_key].append(msg)
                
                # Show duplicates with same content
                for content, msg_list in content_groups.items():
                    if len(msg_list) > 1:
                        duplicate_count += 1
                        print(f"\n{'='*80}")
                        print(f"DUPLICATE #{duplicate_count}: {phone} ({msg_list[0]['recipient_name']})")
                        print(f"Content preview: {content[:50]}...")
                        print(f"\nMessages sent {len(msg_list)} times:")
                        
                        for i, msg in enumerate(msg_list, 1):
                            print(f"\n  Message {i}:")
                            print(f"    ID: {msg['id']}")
                            print(f"    Status: {msg['status']}")
                            print(f"    Created: {msg['created_at']}")
                            print(f"    Scheduled: {msg['scheduled_at']}")
                            print(f"    Sent: {msg['sent_at']}")
                            print(f"    Device: {msg['device_id']}")
                            print(f"    Campaign: {msg['campaign_id']}")
                            print(f"    Sequence: {msg['sequence_id']}")
                            print(f"    Step ID: {msg['sequence_stepid']}")
        
        # Check for timing patterns
        print(f"\n{'='*80}")
        print("\n=== Analyzing Duplicate Patterns ===\n")
        
        # Check messages with very close timestamps
        query2 = """
        SELECT 
            a.id as id1,
            b.id as id2,
            a.recipient_phone,
            a.status as status1,
            b.status as status2,
            a.created_at as created1,
            b.created_at as created2,
            a.sent_at as sent1,
            b.sent_at as sent2,
            TIMESTAMPDIFF(SECOND, a.created_at, b.created_at) as create_diff_seconds,
            TIMESTAMPDIFF(SECOND, a.sent_at, b.sent_at) as sent_diff_seconds
        FROM broadcast_messages a
        JOIN broadcast_messages b ON a.recipient_phone = b.recipient_phone
            AND a.id < b.id
            AND LEFT(a.content, 100) = LEFT(b.content, 100)
        WHERE (DATE(a.created_at) = '2025-08-06' OR DATE(b.created_at) = '2025-08-06')
        ORDER BY create_diff_seconds
        LIMIT 20
        """
        
        cursor.execute(query2)
        close_messages = cursor.fetchall()
        
        print(f"Messages sent very close together:")
        for msg in close_messages:
            print(f"\nPhone: {msg['recipient_phone']}")
            print(f"  IDs: {msg['id1']} vs {msg['id2']}")
            print(f"  Status: {msg['status1']} vs {msg['status2']}")
            print(f"  Created diff: {msg['create_diff_seconds']} seconds")
            print(f"  Sent diff: {msg['sent_diff_seconds']} seconds")
            print(f"  Created: {msg['created1']} vs {msg['created2']}")
            print(f"  Sent: {msg['sent1']} vs {msg['sent2']}")
        
        # Check worker patterns
        print(f"\n{'='*80}")
        print("\n=== Checking Processing Patterns ===\n")
        
        # Messages by status
        query3 = """
        SELECT 
            status,
            COUNT(*) as count,
            MIN(created_at) as first_created,
            MAX(created_at) as last_created
        FROM broadcast_messages
        WHERE DATE(created_at) = '2025-08-06'
        GROUP BY status
        """
        
        cursor.execute(query3)
        status_summary = cursor.fetchall()
        
        print("Messages by status:")
        for row in status_summary:
            print(f"  {row['status']}: {row['count']} messages")
            print(f"    First: {row['first_created']}")
            print(f"    Last: {row['last_created']}")
        
        # Check for messages created at exact same time
        query4 = """
        SELECT 
            created_at,
            COUNT(*) as count,
            GROUP_CONCAT(id ORDER BY id) as message_ids,
            GROUP_CONCAT(DISTINCT recipient_phone) as phones
        FROM broadcast_messages
        WHERE DATE(created_at) = '2025-08-06'
        GROUP BY created_at
        HAVING COUNT(*) > 5
        ORDER BY count DESC
        LIMIT 10
        """
        
        cursor.execute(query4)
        same_time = cursor.fetchall()
        
        print(f"\n\nMessages created at exact same timestamp:")
        for row in same_time:
            print(f"\n  Time: {row['created_at']}")
            print(f"  Count: {row['count']} messages")
            print(f"  IDs: {row['message_ids'][:100]}...")
            print(f"  Phones: {row['phones'][:100]}...")

finally:
    connection.close()

print("\n=== Analysis Complete ===")
