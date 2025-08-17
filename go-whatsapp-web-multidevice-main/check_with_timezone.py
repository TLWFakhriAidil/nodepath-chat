import mysql.connector
from datetime import datetime, timedelta
import sys

sys.stdout.reconfigure(encoding='utf-8')

config = {
    'host': '159.89.198.71',
    'port': 3306,
    'database': 'admin_railway',
    'user': 'admin_aqil',
    'password': 'admin_aqil'
}

try:
    conn = mysql.connector.connect(**config)
    cursor = conn.cursor(dictionary=True)
    
    print("=== CHECKING WITH +8 HOUR TIMEZONE ADJUSTMENT ===\n")
    print("WhatsApp shows: 1:09 PM and 1:10 PM")
    print("Database time would be: 5:09 AM and 5:10 AM UTC\n")
    
    # Check for messages sent around 5:09-5:10 AM UTC (which is 1:09-1:10 PM MYT)
    query1 = """
    SELECT 
        id,
        recipient_phone,
        LEFT(content, 150) as content_preview,
        status,
        DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') as created_at_utc,
        DATE_FORMAT(sent_at, '%Y-%m-%d %H:%i:%s') as sent_at_utc,
        DATE_FORMAT(DATE_ADD(sent_at, INTERVAL 8 HOUR), '%Y-%m-%d %H:%i:%s') as sent_at_myt,
        device_id,
        sequence_stepid,
        processing_worker_id
    FROM broadcast_messages
    WHERE recipient_phone LIKE '%128198574%'
        AND DATE(sent_at) = CURDATE()
        AND TIME(sent_at) BETWEEN '05:00:00' AND '05:20:00'
    ORDER BY sent_at
    """
    
    cursor.execute(query1)
    results = cursor.fetchall()
    
    if results:
        print(f"Found {len(results)} messages sent around that time:\n")
        for msg in results:
            print(f"ID: {msg['id']}")
            print(f"Phone: {msg['recipient_phone']}")
            print(f"Status: {msg['status']}")
            print(f"Created (UTC): {msg['created_at_utc']}")
            print(f"Sent (UTC): {msg['sent_at_utc']}")
            print(f"Sent (MYT +8): {msg['sent_at_myt']}")
            print(f"Worker ID: {msg['processing_worker_id']}")
            print(f"Sequence Step: {msg['sequence_stepid']}")
            print(f"Content: {msg['content_preview']}...")
            print("-" * 80)
    else:
        print("No messages found for 5:00-5:20 AM UTC")
    
    # Check for ANY duplicate content sent today with 1-minute gaps
    print("\n\n=== CHECKING FOR 1-MINUTE GAP DUPLICATES (WITH TIMEZONE) ===\n")
    
    query2 = """
    SELECT 
        a.id as id1,
        b.id as id2,
        a.recipient_phone,
        DATE_FORMAT(a.sent_at, '%H:%i:%s') as sent_time1_utc,
        DATE_FORMAT(b.sent_at, '%H:%i:%s') as sent_time2_utc,
        DATE_FORMAT(DATE_ADD(a.sent_at, INTERVAL 8 HOUR), '%H:%i:%s') as sent_time1_myt,
        DATE_FORMAT(DATE_ADD(b.sent_at, INTERVAL 8 HOUR), '%H:%i:%s') as sent_time2_myt,
        TIMESTAMPDIFF(SECOND, a.sent_at, b.sent_at) as gap_seconds,
        a.sequence_stepid as step1,
        b.sequence_stepid as step2,
        LEFT(a.content, 50) as content1,
        LEFT(b.content, 50) as content2
    FROM broadcast_messages a
    JOIN broadcast_messages b ON a.recipient_phone = b.recipient_phone
    WHERE a.id != b.id
        AND a.sent_at < b.sent_at
        AND TIMESTAMPDIFF(SECOND, a.sent_at, b.sent_at) BETWEEN 30 AND 120
        AND DATE(a.sent_at) = CURDATE()
        AND a.content = b.content
    ORDER BY a.sent_at DESC
    LIMIT 10
    """
    
    cursor.execute(query2)
    gaps = cursor.fetchall()
    
    if gaps:
        print(f"Found {len(gaps)} duplicate messages with ~1 minute gap:\n")
        for gap in gaps:
            print(f"Phone: {gap['recipient_phone']}")
            print(f"Message 1: {gap['id1']} sent at {gap['sent_time1_utc']} UTC ({gap['sent_time1_myt']} MYT)")
            print(f"Message 2: {gap['id2']} sent at {gap['sent_time2_utc']} UTC ({gap['sent_time2_myt']} MYT)")
            print(f"Gap: {gap['gap_seconds']} seconds")
            print(f"Content: {gap['content1']}...")
            print(f"Step IDs: {gap['step1']} vs {gap['step2']}")
            print("-" * 80)
    else:
        print("No duplicate content with 1-minute gaps found")
    
    # Check all messages sent in the last 2 hours
    print("\n\n=== ALL MESSAGES FOR THIS PHONE IN LAST 2 HOURS ===\n")
    
    query3 = """
    SELECT 
        id,
        status,
        DATE_FORMAT(sent_at, '%H:%i:%s') as sent_utc,
        DATE_FORMAT(DATE_ADD(sent_at, INTERVAL 8 HOUR), '%H:%i:%s') as sent_myt,
        sequence_stepid,
        LEFT(content, 80) as preview
    FROM broadcast_messages
    WHERE recipient_phone LIKE '%128198574%'
        AND sent_at >= DATE_SUB(NOW(), INTERVAL 2 HOUR)
        AND sent_at IS NOT NULL
    ORDER BY sent_at DESC
    """
    
    cursor.execute(query3)
    recent = cursor.fetchall()
    
    if recent:
        print(f"Recent messages ({len(recent)} found):\n")
        for msg in recent:
            print(f"ID: {msg['id']}")
            print(f"Sent: {msg['sent_utc']} UTC = {msg['sent_myt']} MYT")
            print(f"Status: {msg['status']}")
            print(f"Preview: {msg['preview']}...")
            print()
    else:
        print("No messages sent in last 2 hours")
        
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
