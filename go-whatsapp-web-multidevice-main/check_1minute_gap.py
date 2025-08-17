import mysql.connector
from datetime import datetime
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
    
    print("=== CHECKING FOR MESSAGES SENT TO 60128198574 BETWEEN 1:00 PM - 1:15 PM TODAY ===\n")
    
    # Check for any messages in that time window
    query1 = """
    SELECT 
        id,
        recipient_phone,
        LEFT(content, 100) as content_preview,
        status,
        DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') as created_at,
        DATE_FORMAT(sent_at, '%Y-%m-%d %H:%i:%s') as sent_at,
        DATE_FORMAT(processing_started_at, '%Y-%m-%d %H:%i:%s') as processing_started,
        device_id,
        sequence_stepid,
        processing_worker_id
    FROM broadcast_messages
    WHERE recipient_phone LIKE '%128198574%'
        AND DATE(COALESCE(sent_at, created_at)) = CURDATE()
        AND (
            (TIME(sent_at) BETWEEN '13:00:00' AND '13:15:00')
            OR (TIME(created_at) BETWEEN '13:00:00' AND '13:15:00')
        )
    ORDER BY COALESCE(sent_at, created_at)
    """
    
    cursor.execute(query1)
    results = cursor.fetchall()
    
    if results:
        print(f"Found {len(results)} messages in this time window:\n")
        for msg in results:
            print(f"ID: {msg['id']}")
            print(f"Phone: {msg['recipient_phone']}")
            print(f"Status: {msg['status']}")
            print(f"Created: {msg['created_at']}")
            print(f"Sent: {msg['sent_at']}")
            print(f"Processing started: {msg['processing_started']}")
            print(f"Worker ID: {msg['processing_worker_id']}")
            print(f"Sequence Step: {msg['sequence_stepid']}")
            print(f"Content: {msg['content_preview']}...")
            print("-" * 80)
    else:
        print("No messages found in database for this time window")
    
    # Check for messages with exact 1-minute gap
    print("\n\n=== CHECKING FOR MESSAGES WITH 1-MINUTE GAPS ===\n")
    
    query2 = """
    SELECT 
        a.id as id1,
        b.id as id2,
        a.recipient_phone,
        DATE_FORMAT(a.sent_at, '%H:%i:%s') as sent_time1,
        DATE_FORMAT(b.sent_at, '%H:%i:%s') as sent_time2,
        TIMESTAMPDIFF(SECOND, a.sent_at, b.sent_at) as gap_seconds,
        a.content = b.content as same_content,
        a.sequence_stepid as step1,
        b.sequence_stepid as step2,
        a.processing_worker_id as worker1,
        b.processing_worker_id as worker2
    FROM broadcast_messages a
    JOIN broadcast_messages b ON a.recipient_phone = b.recipient_phone
    WHERE a.id != b.id
        AND a.sent_at < b.sent_at
        AND TIMESTAMPDIFF(SECOND, a.sent_at, b.sent_at) BETWEEN 30 AND 90
        AND DATE(a.sent_at) = CURDATE()
        AND a.recipient_phone LIKE '%128198574%'
    ORDER BY a.sent_at
    """
    
    cursor.execute(query2)
    gaps = cursor.fetchall()
    
    if gaps:
        print(f"Found {len(gaps)} message pairs with ~1 minute gap:\n")
        for gap in gaps:
            print(f"Message 1 ID: {gap['id1']} sent at {gap['sent_time1']}")
            print(f"Message 2 ID: {gap['id2']} sent at {gap['sent_time2']}")
            print(f"Gap: {gap['gap_seconds']} seconds")
            print(f"Same content: {'YES' if gap['same_content'] else 'NO'}")
            print(f"Step IDs: {gap['step1']} vs {gap['step2']}")
            print(f"Worker IDs: {gap['worker1']} vs {gap['worker2']}")
            print("-" * 80)
    else:
        print("No messages with 1-minute gaps found")
    
    # Check if these are retry attempts
    print("\n\n=== CHECKING ERROR MESSAGES AND RETRIES ===\n")
    
    query3 = """
    SELECT 
        id,
        status,
        error_message,
        DATE_FORMAT(created_at, '%H:%i:%s') as created_time,
        DATE_FORMAT(sent_at, '%H:%i:%s') as sent_time,
        processing_worker_id,
        sequence_stepid
    FROM broadcast_messages
    WHERE recipient_phone LIKE '%128198574%'
        AND DATE(created_at) = CURDATE()
        AND (error_message IS NOT NULL OR status IN ('failed', 'retry'))
    ORDER BY created_at
    """
    
    cursor.execute(query3)
    errors = cursor.fetchall()
    
    if errors:
        print(f"Found {len(errors)} messages with errors/retries:\n")
        for err in errors:
            print(f"ID: {err['id']}")
            print(f"Status: {err['status']}")
            print(f"Error: {err['error_message']}")
            print(f"Times: Created {err['created_time']}, Sent {err['sent_time']}")
            print("-" * 40)
    else:
        print("No error messages or retries found")
        
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
