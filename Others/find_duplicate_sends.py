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
        print("=== ANALYZING DUPLICATE SENDS ===\n")
        
        # Find messages sent to same recipient with same content on Aug 5-6
        query = """
        SELECT 
            a.recipient_phone,
            a.recipient_name,
            COUNT(*) as duplicate_count,
            MIN(a.sent_at) as first_sent,
            MAX(a.sent_at) as last_sent,
            TIMESTAMPDIFF(MINUTE, MIN(a.sent_at), MAX(a.sent_at)) as minutes_apart,
            LEFT(a.content, 50) as content_preview,
            GROUP_CONCAT(DISTINCT a.device_id) as devices_used,
            GROUP_CONCAT(a.id ORDER BY a.sent_at) as message_ids,
            GROUP_CONCAT(a.sent_at ORDER BY a.sent_at) as all_sent_times
        FROM broadcast_messages a
        WHERE a.status = 'sent' 
            AND a.sent_at BETWEEN '2025-08-05 20:00:00' AND '2025-08-06 08:00:00'
        GROUP BY a.recipient_phone, a.recipient_name, LEFT(a.content, 50)
        HAVING COUNT(*) > 1
        ORDER BY minutes_apart ASC
        LIMIT 20
        """
        
        cursor.execute(query)
        duplicates = cursor.fetchall()
        
        print(f"Found {len(duplicates)} cases of duplicate sends\n")
        
        for dup in duplicates:
            print(f"Phone: {dup['recipient_phone']} ({dup['recipient_name']})")
            print(f"  Sent {dup['duplicate_count']} times")
            print(f"  Time apart: {dup['minutes_apart']} minutes")
            print(f"  First: {dup['first_sent']}")
            print(f"  Last: {dup['last_sent']}")
            print(f"  Content: {dup['content_preview']}...")
            print(f"  Devices: {dup['devices_used']}")
            
            # Parse the sent times
            sent_times = dup['all_sent_times'].split(',')
            message_ids = dup['message_ids'].split(',')
            
            print(f"\n  Individual sends:")
            for i, (msg_id, sent_time) in enumerate(zip(message_ids, sent_times)):
                print(f"    {i+1}. ID: {msg_id[:8]}... at {sent_time}")
            print()

finally:
    connection.close()
