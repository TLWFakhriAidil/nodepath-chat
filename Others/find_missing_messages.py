import pymysql
import os
from datetime import datetime, timedelta

# Get MySQL connection
mysql_uri = os.getenv('MYSQL_URI', 'mysql://admin_aqil:admin_aqil@159.89.198.71:3306/admin_railway')

# Parse MySQL URI
uri_parts = mysql_uri.replace('mysql://', '').split('@')
user_pass = uri_parts[0].split(':')
host_db = uri_parts[1].split('/')
host_port = host_db[0].split(':')

connection = pymysql.connect(
    host=host_port[0],
    port=int(host_port[1]),
    user=user_pass[0],
    password=user_pass[1],
    database=host_db[1],
    cursorclass=pymysql.cursors.DictCursor
)

try:
    with connection.cursor() as cursor:
        print("INVESTIGATING THE MISSING MESSAGES")
        print("Looking for messages that match WhatsApp content")
        print("=" * 80)
        
        # Search for messages by content patterns
        search_patterns = [
            ("Pinjam masa Miss", "6:49 AM message"),
            ("Malam Miss", "6:52 AM message"),
            ("Lebih 90%", "1:11/1:13 PM messages")
        ]
        
        for pattern, desc in search_patterns:
            print(f"\nSearching for: {desc} (contains '{pattern}')")
            print("-" * 60)
            
            cursor.execute("""
                SELECT 
                    bm.id,
                    bm.content,
                    bm.status,
                    bm.scheduled_at,
                    bm.sent_at,
                    DATE(bm.scheduled_at) as scheduled_date,
                    DATE(DATE_ADD(bm.sent_at, INTERVAL 8 HOUR)) as sent_date_malaysia,
                    TIME(DATE_ADD(bm.sent_at, INTERVAL 8 HOUR)) as sent_time_malaysia,
                    bm.device_id,
                    ud.device_name,
                    s.name as sequence_name
                FROM broadcast_messages bm
                LEFT JOIN user_devices ud ON ud.id = bm.device_id
                LEFT JOIN sequences s ON s.id = bm.sequence_id
                WHERE bm.recipient_phone = '601117089042'
                AND bm.content LIKE %s
                ORDER BY bm.scheduled_at
            """, (f'%{pattern}%',))
            
            results = cursor.fetchall()
            
            if results:
                print(f"Found {len(results)} message(s):")
                for msg in results:
                    print(f"\n  Message ID: {msg['id'][:8]}...")
                    print(f"  Content: {msg['content'][:80]}...")
                    print(f"  Status: {msg['status']}")
                    print(f"  Scheduled for: {msg['scheduled_date']} at {msg['scheduled_at'].time()}")
                    if msg['sent_at']:
                        print(f"  Actually sent: {msg['sent_date_malaysia']} at {msg['sent_time_malaysia']} (Malaysia)")
                        if msg['scheduled_date'] != msg['sent_date_malaysia']:
                            print(f"  ⚠️  SENT ON DIFFERENT DATE! Scheduled {msg['scheduled_date']} → Sent {msg['sent_date_malaysia']}")
                    print(f"  Device: {msg['device_name']}")
                    print(f"  Sequence: {msg['sequence_name']}")
            else:
                print(f"  NOT FOUND in database!")
                
        # Check if there are messages scheduled for Aug 4 but sent on Aug 5
        print("\n" + "=" * 80)
        print("CHECKING FOR DELAYED MESSAGES:")
        print("Messages scheduled for Aug 4 but possibly sent on Aug 5")
        print("-" * 60)
        
        cursor.execute("""
            SELECT 
                bm.id,
                bm.content,
                bm.status,
                bm.scheduled_at,
                bm.sent_at,
                DATE_ADD(bm.sent_at, INTERVAL 8 HOUR) as sent_malaysia,
                TIMESTAMPDIFF(HOUR, bm.scheduled_at, bm.sent_at) as hours_delay
            FROM broadcast_messages bm
            WHERE bm.recipient_phone = '601117089042'
            AND DATE(bm.scheduled_at) = '2025-08-04'
            AND bm.sent_at IS NOT NULL
            ORDER BY bm.sent_at
        """)
        
        aug4_messages = cursor.fetchall()
        
        for msg in aug4_messages:
            print(f"\nContent: {msg['content'][:60]}...")
            print(f"  Scheduled: {msg['scheduled_at']}")
            print(f"  Sent: {msg['sent_malaysia']} (Malaysia time)")
            print(f"  Delay: {msg['hours_delay']} hours")
            
            # Check if sent on Aug 5 Malaysia time
            if msg['sent_malaysia'].date() == datetime(2025, 8, 5).date():
                print(f"  ⚠️  This was scheduled for Aug 4 but sent on Aug 5!")
                
except Exception as e:
    print(f"Error: {e}")
finally:
    connection.close()
