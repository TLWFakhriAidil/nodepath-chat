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
        print("CHECKING FOR DUPLICATE ENTRIES FOR 601117089042")
        print("=" * 80)
        
        # Check for duplicates by sequence_stepid
        cursor.execute("""
            SELECT 
                bm.id,
                bm.recipient_phone,
                bm.sequence_stepid,
                bm.content,
                bm.status,
                bm.scheduled_at,
                bm.sent_at,
                bm.created_at,
                ss.day,
                ss.content as step_template,
                s.name as sequence_name
            FROM broadcast_messages bm
            LEFT JOIN sequence_steps ss ON ss.id = bm.sequence_stepid
            LEFT JOIN sequences s ON s.id = bm.sequence_id
            WHERE bm.recipient_phone = '601117089042'
            AND bm.sequence_id IS NOT NULL
            ORDER BY bm.sequence_stepid, bm.created_at
        """)
        
        messages = cursor.fetchall()
        
        # Group by sequence_stepid to find duplicates
        by_stepid = {}
        for msg in messages:
            stepid = msg['sequence_stepid']
            if stepid not in by_stepid:
                by_stepid[stepid] = []
            by_stepid[stepid].append(msg)
        
        print("DUPLICATE CHECK BY SEQUENCE STEP:")
        print("-" * 60)
        
        for stepid, msgs in by_stepid.items():
            if len(msgs) > 1:
                print(f"\n⚠️  DUPLICATE FOUND! Step ID: {stepid}")
                print(f"   Sequence: {msgs[0]['sequence_name']}")
                print(f"   Step template: {msgs[0]['step_template'][:50] if msgs[0]['step_template'] else 'No template'}...")
                print(f"   Number of duplicates: {len(msgs)}")
                
                for i, msg in enumerate(msgs):
                    print(f"\n   Duplicate {i+1}:")
                    print(f"     ID: {msg['id'][:8]}...")
                    print(f"     Status: {msg['status']}")
                    print(f"     Created: {msg['created_at']}")
                    print(f"     Scheduled: {msg['scheduled_at']}")
                    print(f"     Sent: {msg['sent_at'] if msg['sent_at'] else 'Not sent'}")
                    print(f"     Content preview: {msg['content'][:60] if msg['content'] else 'No content'}...")
            else:
                print(f"\nStep ID: {stepid} - OK (no duplicates)")
                
        # Check for similar content that might be duplicates with variation
        print("\n" + "=" * 80)
        print("CHECKING FOR SIMILAR CONTENT (VARIATIONS):")
        print("-" * 60)
        
        # Look for messages with similar patterns
        cursor.execute("""
            SELECT 
                bm.id,
                bm.content,
                bm.scheduled_at,
                bm.sent_at,
                bm.status,
                bm.sequence_stepid,
                DATE_ADD(bm.sent_at, INTERVAL 8 HOUR) as sent_malaysia_time
            FROM broadcast_messages bm
            WHERE bm.recipient_phone = '601117089042'
            AND bm.content LIKE '%first time nak cari solusi%'
            ORDER BY bm.scheduled_at
        """)
        
        similar = cursor.fetchall()
        
        if similar:
            print(f"\nFound {len(similar)} messages with 'first time nak cari solusi':")
            for msg in similar:
                print(f"\n  ID: {msg['id'][:8]}...")
                print(f"  Content: {msg['content'][:80]}...")
                print(f"  Status: {msg['status']}")
                print(f"  Scheduled: {msg['scheduled_at']}")
                if msg['sent_at']:
                    print(f"  Sent (Malaysia): {msg['sent_malaysia_time']}")
                print(f"  Step ID: {msg['sequence_stepid']}")
                
        # Check if there's a greeting variation pattern
        print("\n" + "=" * 80)
        print("ANALYSIS: Why might greetings vary?")
        print("-" * 60)
        
        # Check if messages are scheduled at different times of day
        cursor.execute("""
            SELECT 
                HOUR(scheduled_at) as hour,
                COUNT(*) as count,
                GROUP_CONCAT(DISTINCT LEFT(content, 20)) as greeting_samples
            FROM broadcast_messages
            WHERE recipient_phone = '601117089042'
            AND sequence_id IS NOT NULL
            GROUP BY HOUR(scheduled_at)
            ORDER BY hour
        """)
        
        hourly = cursor.fetchall()
        
        print("\nMessages by hour (might explain greeting variations):")
        for h in hourly:
            print(f"  Hour {h['hour']:02d}:00 - {h['count']} messages")
            if h['greeting_samples']:
                print(f"    Greetings: {h['greeting_samples']}")
                
except Exception as e:
    print(f"Error: {e}")
finally:
    connection.close()
