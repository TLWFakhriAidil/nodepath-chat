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
        print("SPECIFIC MESSAGES SENT ON AUGUST 5, 2025")
        print("Recipient: 601117089042")
        print("=" * 80)
        
        # Get the 2 messages that were sent on Aug 5
        cursor.execute("""
            SELECT 
                bm.id,
                bm.content,
                bm.media_url,
                bm.sent_at,
                DATE_ADD(bm.sent_at, INTERVAL 8 HOUR) as malaysia_time,
                TIME(DATE_ADD(bm.sent_at, INTERVAL 8 HOUR)) as malaysia_time_only
            FROM broadcast_messages bm
            WHERE bm.recipient_phone = '601117089042'
            AND bm.status = 'sent'
            AND bm.sent_at >= '2025-08-04 16:00:00'
            AND bm.sent_at < '2025-08-05 16:00:00'
            ORDER BY bm.sent_at
        """)
        
        messages = cursor.fetchall()
        
        print(f"DATABASE SHOWS {len(messages)} MESSAGES SENT ON AUG 5:\n")
        
        for i, msg in enumerate(messages):
            print(f"MESSAGE {i+1}:")
            print(f"  Time sent (Malaysia): {msg['malaysia_time_only']}")
            print(f"  Content: {msg['content']}")
            print(f"  Has image: {'YES' if msg['media_url'] else 'NO'}")
            print("-" * 60)
            
        print("\n\nWHATSAPP SCREENSHOT SHOWS 4 MESSAGES:")
        print("1. 6:49 AM - 'Pinjam masa Miss, Akak first time nak cari solusi...'")
        print("2. 6:52 AM - 'Malam Miss, Akak first time nak cari solusi...'") 
        print("3. 1:11 PM - [Image] 'Assalamualaikum Miss, Lebih 90% anak2...'")
        print("4. 1:13 PM - [Image] 'Assalamualaikum Miss, Lebih 90% anak2... baru nak bertindak...'")
        
        print("\n" + "=" * 80)
        print("COMPARISON:")
        print("Database has: 2 messages")
        print("WhatsApp has: 4 messages")
        print("MISSING: 2 messages (6:49 AM and 1:11 PM)")
        
except Exception as e:
    print(f"Error: {e}")
finally:
    connection.close()
