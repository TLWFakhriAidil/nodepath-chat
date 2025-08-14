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
        print("Checking messages sent on August 5, 2025 (Malaysia time)")
        print("For recipient: 601117089042")
        print("=" * 80)
        
        # Check all messages sent between Aug 4 16:00 (server) to Aug 5 16:00 (server)
        # This covers Aug 5 00:00 to Aug 5 24:00 Malaysia time
        cursor.execute("""
            SELECT 
                bm.id,
                bm.status,
                bm.scheduled_at,
                bm.sent_at,
                bm.content,
                bm.media_url,
                bm.device_id,
                s.name as sequence_name,
                ud.device_name,
                bm.created_at,
                TIME(DATE_ADD(bm.sent_at, INTERVAL 8 HOUR)) as malaysia_sent_time,
                DATE(DATE_ADD(bm.sent_at, INTERVAL 8 HOUR)) as malaysia_sent_date
            FROM broadcast_messages bm
            LEFT JOIN sequences s ON s.id = bm.sequence_id
            LEFT JOIN user_devices ud ON ud.id = bm.device_id
            WHERE bm.recipient_phone = '601117089042'
            AND bm.status = 'sent'
            AND bm.sent_at >= '2025-08-04 16:00:00'
            AND bm.sent_at < '2025-08-05 16:00:00'
            ORDER BY bm.sent_at
        """)
        
        messages = cursor.fetchall()
        
        print(f"Messages sent on August 5 (Malaysia time): {len(messages)}\n")
        
        # Match with WhatsApp timestamps
        whatsapp_times = ['6:49 am', '6:52 am', '1:11 pm', '1:13 pm']
        
        print("Matching with WhatsApp screenshot times:")
        print("-" * 60)
        
        for i, msg in enumerate(messages):
            print(f"\nMessage {i+1}:")
            print(f"  WhatsApp shows: ~{whatsapp_times[i] if i < len(whatsapp_times) else 'Unknown'}")
            print(f"  Our record shows: {msg['malaysia_sent_time']}")
            print(f"  Content: {msg['content'][:60]}...")
            print(f"  Has image: {'Yes' if msg['media_url'] else 'No'}")
            print(f"  From: {msg['sequence_name']}")
            print(f"  Device: {msg['device_name']}")
            
        # Check for any duplicate sends or resends
        print("\n" + "=" * 80)
        print("Checking for duplicate/resend patterns:")
        
        cursor.execute("""
            SELECT 
                content,
                COUNT(*) as count,
                GROUP_CONCAT(TIME(DATE_ADD(sent_at, INTERVAL 8 HOUR)) ORDER BY sent_at) as sent_times
            FROM broadcast_messages
            WHERE recipient_phone = '601117089042'
            AND status = 'sent'
            AND sent_at >= '2025-08-04 16:00:00'
            AND sent_at < '2025-08-05 16:00:00'
            GROUP BY content
            HAVING count > 1
        """)
        
        duplicates = cursor.fetchall()
        
        if duplicates:
            print("\nDUPLICATES FOUND:")
            for dup in duplicates:
                print(f"  Content: {dup['content'][:50]}...")
                print(f"  Sent {dup['count']} times at: {dup['sent_times']}")
        else:
            print("\nNo exact duplicates found.")
            
        # Check if messages from different dates were sent on same day
        print("\n" + "=" * 80)
        print("Checking original scheduled dates vs actual sent dates:")
        
        for msg in messages:
            scheduled_date = msg['scheduled_at'].date()
            sent_date = msg['malaysia_sent_date']
            if scheduled_date != sent_date:
                print(f"\nMISMATCH FOUND:")
                print(f"  Content: {msg['content'][:50]}...")
                print(f"  Was scheduled for: {scheduled_date}")
                print(f"  Actually sent on: {sent_date}")
                print(f"  Sent at: {msg['malaysia_sent_time']}")
                
except Exception as e:
    print(f"Error: {e}")
finally:
    connection.close()
