import pymysql
from datetime import datetime, timedelta

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
        print("=== ANALYZING EXACT DUPLICATE TIMING ===\n")
        
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
            LEFT(a.content, 50) as content_preview,
            TIMESTAMPDIFF(SECOND, a.sent_at, b.sent_at) as sent_diff_seconds
        FROM broadcast_messages a
        JOIN broadcast_messages b ON 
            a.recipient_phone = b.recipient_phone
            AND a.id < b.id
            AND a.sequence_stepid = b.sequence_stepid
            AND LEFT(a.content, 100) = LEFT(b.content, 100)
        WHERE DATE(a.sent_at) = '2025-08-06'
            AND a.status = 'sent'
            AND b.status = 'sent'
        ORDER BY sent_diff_seconds ASC
        LIMIT 20
        """
        
        cursor.execute(query)
        duplicates = cursor.fetchall()
        
        print(f"Found {len(duplicates)} duplicate pairs sent on same day\n")
        
        for dup in duplicates:
            print(f"Phone: {dup['recipient_phone']} ({dup['recipient_name']})")
            print(f"  Same Step ID: {dup['step1']}")
            print(f"  Message 1: {dup['id1'][:8]}... on device {dup['device1'][:8]}...")
            print(f"  Message 2: {dup['id2'][:8]}... on device {dup['device2'][:8]}...")
            print(f"  Sent times: {dup['sent1']} vs {dup['sent2']}")
            print(f"  Time diff: {dup['sent_diff_seconds']} seconds")
            print(f"  Content: {dup['content_preview']}...")
            print()
        
        # Check if it's same device or different devices
        print("\n=== DEVICE ANALYSIS ===\n")
        
        query = """
        SELECT 
            CASE 
                WHEN a.device_id = b.device_id THEN 'SAME_DEVICE'
                ELSE 'DIFFERENT_DEVICES'
            END as device_pattern,
            COUNT(*) as count
        FROM broadcast_messages a
        JOIN broadcast_messages b ON 
            a.recipient_phone = b.recipient_phone
            AND a.id < b.id
            AND a.sequence_stepid = b.sequence_stepid
            AND LEFT(a.content, 100) = LEFT(b.content, 100)
        WHERE DATE(a.sent_at) = '2025-08-06'
            AND a.status = 'sent'
            AND b.status = 'sent'
        GROUP BY device_pattern
        """
        
        cursor.execute(query)
        device_patterns = cursor.fetchall()
        
        for pattern in device_patterns:
            print(f"{pattern['device_pattern']}: {pattern['count']} cases")
        
        # Check creation time patterns
        print("\n=== CREATION TIME ANALYSIS ===\n")
        
        query = """
        SELECT 
            a.recipient_phone,
            a.created_at as created1,
            b.created_at as created2,
            TIMESTAMPDIFF(SECOND, a.created_at, b.created_at) as create_diff,
            a.id as id1,
            b.id as id2
        FROM broadcast_messages a
        JOIN broadcast_messages b ON 
            a.recipient_phone = b.recipient_phone
            AND a.id < b.id
            AND a.sequence_stepid = b.sequence_stepid
        WHERE DATE(a.sent_at) = '2025-08-06'
            AND a.status = 'sent'
            AND b.status = 'sent'
        ORDER BY create_diff ASC
        LIMIT 10
        """
        
        cursor.execute(query)
        creation_times = cursor.fetchall()
        
        for ct in creation_times:
            print(f"Phone: {ct['recipient_phone']}")
            print(f"  Created at EXACT SAME TIME: {ct['created1']}")
            print(f"  Time diff: {ct['create_diff']} seconds")
            print(f"  IDs: {ct['id1'][:8]}... and {ct['id2'][:8]}...")
            print()

finally:
    connection.close()
