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
        print("=== INVESTIGATING EXACT DUPLICATE TIMING ===\n")
        
        # Find messages sent at EXACTLY the same time with same content
        query = """
        SELECT 
            a.id as id1,
            b.id as id2,
            a.recipient_phone,
            a.recipient_name,
            a.device_id as device1,
            b.device_id as device2,
            a.status as status1,
            b.status as status2,
            a.created_at,
            a.scheduled_at,
            a.sent_at as sent1,
            b.sent_at as sent2,
            a.sequence_stepid as step1,
            b.sequence_stepid as step2,
            TIMESTAMPDIFF(SECOND, a.sent_at, b.sent_at) as sent_diff_seconds,
            LEFT(a.content, 50) as content_preview
        FROM broadcast_messages a
        JOIN broadcast_messages b ON 
            a.recipient_phone = b.recipient_phone
            AND a.id < b.id
            AND a.sequence_stepid = b.sequence_stepid  -- Same step!
            AND LEFT(a.content, 200) = LEFT(b.content, 200)  -- Same content!
        WHERE 
            a.status = 'sent' AND b.status = 'sent'
            AND DATE(a.sent_at) = '2025-08-06'
            AND ABS(TIMESTAMPDIFF(SECOND, a.sent_at, b.sent_at)) < 300  -- Within 5 minutes
        ORDER BY sent_diff_seconds
        LIMIT 20
        """
        
        cursor.execute(query)
        exact_duplicates = cursor.fetchall()
        
        print(f"Found {len(exact_duplicates)} exact duplicates sent within 5 minutes\n")
        
        for dup in exact_duplicates:
            print(f"DUPLICATE FOUND:")
            print(f"  Phone: {dup['recipient_phone']} ({dup['recipient_name']})")
            print(f"  Content: {dup['content_preview']}...")
            print(f"  Same Step ID: {dup['step1']}")
            print(f"  Message IDs: {dup['id1']} vs {dup['id2']}")
            print(f"  Devices: {dup['device1']} vs {dup['device2']}")
            print(f"  Sent times: {dup['sent1']} vs {dup['sent2']}")
            print(f"  Time diff: {dup['sent_diff_seconds']} seconds")
            print(f"  Created at: {dup['created_at']}")
            print(f"  Scheduled at: {dup['scheduled_at']}")
            print()
        
        # Check if same device or different devices
        print("\n=== DEVICE PATTERN ANALYSIS ===\n")
        
        query = """
        SELECT 
            CASE 
                WHEN a.device_id = b.device_id THEN 'SAME_DEVICE'
                ELSE 'DIFFERENT_DEVICES'
            END as device_pattern,
            COUNT(*) as count,
            AVG(ABS(TIMESTAMPDIFF(SECOND, a.sent_at, b.sent_at))) as avg_time_diff
        FROM broadcast_messages a
        JOIN broadcast_messages b ON 
            a.recipient_phone = b.recipient_phone
            AND a.id < b.id
            AND a.sequence_stepid = b.sequence_stepid
            AND LEFT(a.content, 200) = LEFT(b.content, 200)
        WHERE 
            a.status = 'sent' AND b.status = 'sent'
            AND DATE(a.sent_at) = '2025-08-06'
        GROUP BY device_pattern
        """
        
        cursor.execute(query)
        device_patterns = cursor.fetchall()
        
        for pattern in device_patterns:
            print(f"{pattern['device_pattern']}: {pattern['count']} duplicates")
            print(f"  Average time difference: {pattern['avg_time_diff']:.1f} seconds")
        
        # Check worker processing pattern
        print("\n\n=== WORKER PROCESSING PATTERN ===\n")
        
        # Messages that were pending at the same time
        query = """
        SELECT 
            DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:00') as minute,
            COUNT(*) as messages_created,
            COUNT(DISTINCT device_id) as devices_used,
            GROUP_CONCAT(DISTINCT status) as statuses
        FROM broadcast_messages
        WHERE DATE(scheduled_at) = '2025-08-06'
            AND sequence_stepid IN (
                SELECT DISTINCT sequence_stepid 
                FROM broadcast_messages 
                WHERE DATE(sent_at) = '2025-08-06'
                GROUP BY recipient_phone, sequence_stepid
                HAVING COUNT(*) > 1
            )
        GROUP BY minute
        HAVING messages_created > 10
        ORDER BY messages_created DESC
        LIMIT 10
        """
        
        cursor.execute(query)
        patterns = cursor.fetchall()
        
        print("Times when duplicate messages were created in bulk:")
        for p in patterns:
            print(f"  {p['minute']}: {p['messages_created']} messages on {p['devices_used']} devices")
            print(f"    Statuses: {p['statuses']}")
        
        # Check specific example timeline
        print("\n\n=== TIMELINE FOR SPECIFIC DUPLICATE ===\n")
        
        example_phone = '60122712014'
        query = """
        SELECT 
            id,
            status,
            device_id,
            created_at,
            scheduled_at,
            sent_at,
            TIMESTAMPDIFF(SECOND, scheduled_at, sent_at) as process_delay,
            sequence_stepid,
            LEFT(content, 50) as content
        FROM broadcast_messages
        WHERE recipient_phone = %s
            AND DATE(scheduled_at) = '2025-08-06'
            AND sequence_stepid = '72e96e33-2169-4d72-8d2d-041eab647e53'
        ORDER BY created_at, sent_at
        """
        
        cursor.execute(query, (example_phone,))
        timeline = cursor.fetchall()
        
        print(f"Timeline for {example_phone}:")
        for msg in timeline:
            print(f"\n  Message ID: {msg['id']}")
            print(f"  Status: {msg['status']}")
            print(f"  Device: {msg['device_id']}")
            print(f"  Created: {msg['created_at']}")
            print(f"  Scheduled: {msg['scheduled_at']}")
            print(f"  Sent: {msg['sent_at']}")
            if msg['sent_at']:
                print(f"  Process delay: {msg['process_delay']} seconds after scheduled time")

finally:
    connection.close()

print("\n=== Analysis Complete ===")
