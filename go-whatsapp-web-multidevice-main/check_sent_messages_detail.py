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
        print("Checking ALL SENT messages for recipient: 601117089042")
        print("Focusing on messages sent today (August 5, 2025)")
        print("=" * 80)
        
        # Get current server time
        cursor.execute("SELECT NOW() as server_time")
        result = cursor.fetchone()
        print(f"Current server time: {result['server_time']}")
        print()
        
        # Check all SENT messages for this recipient
        cursor.execute("""
            SELECT 
                bm.id,
                bm.status,
                bm.scheduled_at,
                bm.sent_at,
                bm.content,
                bm.media_url,
                bm.device_id,
                bm.sequence_id,
                bm.campaign_id,
                s.name as sequence_name,
                c.title as campaign_name,
                ud.device_name,
                bm.created_at
            FROM broadcast_messages bm
            LEFT JOIN sequences s ON s.id = bm.sequence_id
            LEFT JOIN campaigns c ON c.id = bm.campaign_id
            LEFT JOIN user_devices ud ON ud.id = bm.device_id
            WHERE bm.recipient_phone = '601117089042'
            AND bm.status = 'sent'
            ORDER BY bm.sent_at DESC
        """)
        
        sent_messages = cursor.fetchall()
        
        print(f"Total SENT messages for this recipient: {len(sent_messages)}\n")
        
        # Group by sent date
        print("Messages by SENT date (not scheduled date):")
        by_sent_date = {}
        for msg in sent_messages:
            if msg['sent_at']:
                # Add 8 hours to convert to Malaysia time
                malaysia_sent_time = msg['sent_at'] + timedelta(hours=8)
                date_key = malaysia_sent_time.date()
                if date_key not in by_sent_date:
                    by_sent_date[date_key] = []
                by_sent_date[date_key].append(msg)
        
        for date, msgs in sorted(by_sent_date.items(), reverse=True):
            print(f"\n{date} (Malaysia time): {len(msgs)} messages sent")
            
        # Show all sent messages with details
        print("\n" + "=" * 80)
        print("DETAILED SENT MESSAGES:")
        print("=" * 80)
        
        for i, msg in enumerate(sent_messages):
            print(f"\nMessage {i+1}:")
            print(f"  ID: {msg['id'][:8]}...")
            print(f"  From: {msg['sequence_name'] or msg['campaign_name'] or 'Direct'}")
            print(f"  Device: {msg['device_name']} ({msg['device_id'][:8]}...)")
            print(f"  Content: {msg['content'][:50]}..." if msg['content'] else "  Content: [No text/Image only]")
            print(f"  Has media: {'Yes' if msg['media_url'] else 'No'}")
            print(f"  Created at: {msg['created_at']}")
            print(f"  Scheduled for: {msg['scheduled_at']}")
            print(f"  Actually sent: {msg['sent_at']}")
            
            # Calculate Malaysia time
            if msg['sent_at']:
                malaysia_sent = msg['sent_at'] + timedelta(hours=8)
                print(f"  Sent (Malaysia time): {malaysia_sent}")
            
        # Check for duplicate content
        print("\n" + "=" * 80)
        print("CHECKING FOR DUPLICATE MESSAGES:")
        
        # Group by content to find duplicates
        content_groups = {}
        for msg in sent_messages:
            content_key = msg['content'][:50] if msg['content'] else "[Image]"
            if content_key not in content_groups:
                content_groups[content_key] = []
            content_groups[content_key].append(msg)
        
        for content, msgs in content_groups.items():
            if len(msgs) > 1:
                print(f"\nDuplicate found: '{content}...'")
                print(f"  Sent {len(msgs)} times:")
                for msg in msgs:
                    print(f"    - {msg['sent_at']} from {msg['device_name']}")
                    
except Exception as e:
    print(f"Error: {e}")
finally:
    connection.close()
