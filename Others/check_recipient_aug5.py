import pymysql
import os
from datetime import datetime

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
        print("Checking ALL messages scheduled for recipient: 601117089042")
        print("=" * 80)
        
        # Check all messages for this recipient scheduled for any date
        cursor.execute("""
            SELECT 
                bm.id,
                bm.status,
                bm.scheduled_at,
                bm.sent_at,
                bm.content,
                bm.media_url,
                bm.error_message,
                bm.device_id,
                bm.sequence_id,
                bm.campaign_id,
                s.name as sequence_name,
                c.title as campaign_name,
                ud.device_name
            FROM broadcast_messages bm
            LEFT JOIN sequences s ON s.id = bm.sequence_id
            LEFT JOIN campaigns c ON c.id = bm.campaign_id
            LEFT JOIN user_devices ud ON ud.id = bm.device_id
            WHERE bm.recipient_phone = '601117089042'
            ORDER BY bm.scheduled_at
        """)
        
        all_messages = cursor.fetchall()
        
        print(f"Total messages for this recipient: {len(all_messages)}\n")
        
        # Group by date
        by_date = {}
        for msg in all_messages:
            date_key = msg['scheduled_at'].date()
            if date_key not in by_date:
                by_date[date_key] = []
            by_date[date_key].append(msg)
        
        print("Messages grouped by scheduled date:")
        for date, msgs in sorted(by_date.items()):
            print(f"\n{date}: {len(msgs)} messages")
            for i, msg in enumerate(msgs):
                print(f"  {i+1}. Time: {msg['scheduled_at'].time()} | Status: {msg['status']} | {msg['sequence_name'] or msg['campaign_name'] or 'No source'}")
        
        messages = cursor.fetchall()
        
        print(f"Found {len(messages)} messages for this recipient on Aug 5\n")
        
        for i, msg in enumerate(messages):
            print(f"Message {i+1}:")
            print(f"  ID: {msg['id'][:8]}...")
            print(f"  Status: {msg['status']}")
            print(f"  Scheduled: {msg['scheduled_at']}")
            if msg['sent_at']:
                print(f"  Sent at: {msg['sent_at']}")
            print(f"  Device: {msg['device_name']} ({msg['device_id'][:8]}...)")
            
            if msg['sequence_name']:
                print(f"  From Sequence: {msg['sequence_name']}")
            elif msg['campaign_name']:
                print(f"  From Campaign: {msg['campaign_name']}")
                
            print(f"  Content: {msg['content'][:100]}..." if msg['content'] else "  Content: [No text content]")
            if msg['media_url']:
                print(f"  Has media: Yes")
            if msg['error_message']:
                print(f"  Error: {msg['error_message']}")
            print()
        
        # Check status breakdown
        print("\nStatus Summary:")
        cursor.execute("""
            SELECT 
                status,
                COUNT(*) as count
            FROM broadcast_messages
            WHERE recipient_phone = '601117089042'
            AND DATE(scheduled_at) = '2025-08-05'
            GROUP BY status
        """)
        
        statuses = cursor.fetchall()
        for status in statuses:
            print(f"  {status['status']}: {status['count']} messages")
            
        # Check scheduled times
        print("\nScheduled times:")
        cursor.execute("""
            SELECT 
                TIME(scheduled_at) as time,
                COUNT(*) as count
            FROM broadcast_messages
            WHERE recipient_phone = '601117089042'
            AND DATE(scheduled_at) = '2025-08-05'
            GROUP BY TIME(scheduled_at)
            ORDER BY time
        """)
        
        times = cursor.fetchall()
        for t in times:
            print(f"  {t['time']}: {t['count']} messages")
            
except Exception as e:
    print(f"Error: {e}")
finally:
    connection.close()
