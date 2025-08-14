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
        print("=== SEARCHING FOR RACE CONDITION EVIDENCE ===\n")
        
        # Look for messages created/sent within seconds of each other
        query = """
        SELECT 
            recipient_phone,
            COUNT(*) as msg_count,
            MIN(sent_at) as first_sent,
            MAX(sent_at) as last_sent,
            TIMESTAMPDIFF(SECOND, MIN(sent_at), MAX(sent_at)) as time_span,
            GROUP_CONCAT(id ORDER BY sent_at) as message_ids,
            GROUP_CONCAT(status ORDER BY sent_at) as statuses
        FROM broadcast_messages
        WHERE sent_at BETWEEN '2025-08-06 00:00:00' AND '2025-08-06 23:59:59'
        GROUP BY recipient_phone, LEFT(content, 50)
        HAVING COUNT(*) > 1 AND time_span < 60
        ORDER BY time_span ASC
        LIMIT 10
        """
        
        cursor.execute(query)
        results = cursor.fetchall()
        
        print(f"Messages sent within 60 seconds to same recipient:\n")
        
        for r in results:
            print(f"Phone: {r['recipient_phone']}")
            print(f"  Messages: {r['msg_count']}")
            print(f"  Time span: {r['time_span']} seconds")
            print(f"  First sent: {r['first_sent']}")
            print(f"  Last sent: {r['last_sent']}")
            print(f"  Message IDs: {r['message_ids'][:50]}...")
            print()

finally:
    connection.close()
