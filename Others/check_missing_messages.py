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
        print("Checking ALL messages (all statuses) for recipient: 601117089042")
        print("Looking for any message that could have been sent on Aug 5")
        print("=" * 80)
        
        # Check ALL messages regardless of status
        cursor.execute("""
            SELECT 
                bm.id,
                bm.status,
                bm.scheduled_at,
                bm.sent_at,
                bm.content,
                bm.media_url,
                bm.error_message,
                s.name as sequence_name,
                bm.created_at,
                bm.updated_at
            FROM broadcast_messages bm
            LEFT JOIN sequences s ON s.id = bm.sequence_id
            WHERE bm.recipient_phone = '601117089042'
            ORDER BY bm.scheduled_at
        """)
        
        all_messages = cursor.fetchall()
        
        print(f"Total messages for this recipient: {len(all_messages)}\n")
        
        # Group by status
        by_status = {}
        for msg in all_messages:
            status = msg['status']
            if status not in by_status:
                by_status[status] = []
            by_status[status].append(msg)
        
        print("Messages by status:")
        for status, msgs in by_status.items():
            print(f"  {status}: {len(msgs)} messages")
        
        # Look for messages that match WhatsApp content
        print("\n" + "=" * 80)
        print("Matching WhatsApp messages with database records:")
        print("-" * 60)
        
        whatsapp_messages = [
            {"time": "6:49 am", "content": "Pinjam masa Miss, Akak first time nak cari solusi"},
            {"time": "6:52 am", "content": "Malam Miss, Akak first time nak cari solusi"},
            {"time": "1:11 pm", "content": "[Image] Assalamualaikum Miss, Lebih 90% anak2 yang kurang nutrisi"},
            {"time": "1:13 pm", "content": "[Image] Assalamualaikum Miss, Lebih 90% anak2... baru nak bertindak"}
        ]
        
        for wa_msg in whatsapp_messages:
            print(f"\nWhatsApp {wa_msg['time']}: {wa_msg['content'][:50]}...")
            found = False
            
            for msg in all_messages:
                if msg['content'] and wa_msg['content'][8:40] in msg['content']:
                    print(f"  FOUND in DB:")
                    print(f"    Status: {msg['status']}")
                    print(f"    Scheduled: {msg['scheduled_at']}")
                    print(f"    Sent: {msg['sent_at']}")
                    print(f"    Sequence: {msg['sequence_name']}")
                    found = True
                    break
                    
            if not found:
                print(f"  NOT FOUND in database!")
                
        # Check if there are multiple devices sending to same number
        print("\n" + "=" * 80)
        print("Checking devices that sent to this number:")
        
        cursor.execute("""
            SELECT DISTINCT
                bm.device_id,
                ud.device_name,
                COUNT(*) as message_count
            FROM broadcast_messages bm
            LEFT JOIN user_devices ud ON ud.id = bm.device_id
            WHERE bm.recipient_phone = '601117089042'
            AND bm.status = 'sent'
            GROUP BY bm.device_id, ud.device_name
        """)
        
        devices = cursor.fetchall()
        
        for device in devices:
            print(f"\nDevice: {device['device_name']} ({device['device_id'][:8]}...)")
            print(f"  Sent {device['message_count']} messages")
            
        # Check for any resend operations
        print("\n" + "=" * 80)
        print("Checking for messages that were updated (possible resends):")
        
        cursor.execute("""
            SELECT 
                id,
                content,
                status,
                created_at,
                updated_at,
                TIMESTAMPDIFF(MINUTE, created_at, updated_at) as minutes_between_updates
            FROM broadcast_messages
            WHERE recipient_phone = '601117089042'
            AND updated_at > created_at
            AND TIMESTAMPDIFF(MINUTE, created_at, updated_at) > 5
            ORDER BY updated_at
        """)
        
        updated = cursor.fetchall()
        
        if updated:
            print(f"\nFound {len(updated)} messages that were updated after creation:")
            for msg in updated:
                print(f"\n  Content: {msg['content'][:50]}...")
                print(f"  Created: {msg['created_at']}")
                print(f"  Updated: {msg['updated_at']} ({msg['minutes_between_updates']} minutes later)")
                print(f"  Status: {msg['status']}")
                
except Exception as e:
    print(f"Error: {e}")
finally:
    connection.close()
