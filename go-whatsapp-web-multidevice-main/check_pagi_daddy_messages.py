import mysql.connector
from datetime import datetime
import sys

# Set UTF-8 encoding for output
sys.stdout.reconfigure(encoding='utf-8')

# Database connection
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
    print("Connected to MySQL database successfully!")
    
    # Check for messages that contain the text from the WhatsApp screenshot
    print("\n=== SEARCHING FOR MESSAGES WITH 'Pagi Daddy Dassler' ===")
    query1 = """
    SELECT 
        id,
        recipient_phone,
        LEFT(content, 200) as message_preview,
        status,
        DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') as created_at,
        DATE_FORMAT(sent_at, '%Y-%m-%d %H:%i:%s') as sent_at,
        device_id,
        sequence_id,
        sequence_stepid
    FROM broadcast_messages
    WHERE (recipient_phone = '+60128198574' 
        OR recipient_phone = '60128198574'
        OR recipient_phone = '0128198574'
        OR recipient_phone LIKE '%128198574%')
        AND content LIKE '%Pagi Daddy Dassler%'
    ORDER BY created_at DESC
    """
    
    cursor.execute(query1)
    results = cursor.fetchall()
    
    if results:
        print(f"\nFound {len(results)} messages with 'Pagi Daddy Dassler':")
        for i, msg in enumerate(results, 1):
            print(f"\n--- Message {i} ---")
            print(f"ID: {msg['id']}")
            print(f"Phone: {msg['recipient_phone']}")
            print(f"Created: {msg['created_at']}")
            print(f"Sent: {msg['sent_at']}")
            print(f"Status: {msg['status']}")
            print(f"Device: {msg['device_id']}")
            print(f"Sequence Step: {msg['sequence_stepid']}")
            try:
                print(f"Preview: {msg['message_preview']}")
            except:
                print("Preview: [Contains special characters]")
    else:
        print("No messages found with 'Pagi Daddy Dassler'")
    
    # Check messages by time (1:38 PM = 13:38, but also check surrounding times)
    print("\n\n=== MESSAGES SENT BETWEEN 1:30 PM AND 1:45 PM ===")
    query2 = """
    SELECT 
        id,
        recipient_phone,
        status,
        DATE_FORMAT(sent_at, '%H:%i:%s') as sent_time,
        device_id,
        sequence_stepid,
        LENGTH(content) as msg_length
    FROM broadcast_messages
    WHERE (recipient_phone = '+60128198574' 
        OR recipient_phone = '60128198574'
        OR recipient_phone = '0128198574'
        OR recipient_phone LIKE '%128198574%')
        AND DATE(sent_at) = '2025-08-10'
        AND TIME(sent_at) BETWEEN '13:30:00' AND '13:45:00'
    ORDER BY sent_at
    """
    
    cursor.execute(query2)
    time_results = cursor.fetchall()
    
    if time_results:
        print(f"\nFound {len(time_results)} messages sent between 1:30 PM and 1:45 PM:")
        for msg in time_results:
            print(f"\nID: {msg['id']}")
            print(f"Sent at: {msg['sent_time']}")
            print(f"Status: {msg['status']}")
            print(f"Message length: {msg['msg_length']} characters")
            print(f"Sequence Step: {msg['sequence_stepid']}")
    
    # Check for exact duplicates (same sequence_stepid)
    print("\n\n=== CHECKING FOR DUPLICATE SEQUENCE STEPS ===")
    query3 = """
    SELECT 
        sequence_stepid,
        COUNT(*) as count,
        GROUP_CONCAT(id) as message_ids,
        GROUP_CONCAT(DATE_FORMAT(sent_at, '%Y-%m-%d %H:%i:%s')) as sent_times,
        GROUP_CONCAT(status) as statuses
    FROM broadcast_messages
    WHERE (recipient_phone = '+60128198574' 
        OR recipient_phone = '60128198574'
        OR recipient_phone = '0128198574'
        OR recipient_phone LIKE '%128198574%')
        AND sequence_stepid IS NOT NULL
        AND DATE(created_at) >= '2025-08-09'
    GROUP BY sequence_stepid
    HAVING COUNT(*) > 1
    """
    
    cursor.execute(query3)
    step_duplicates = cursor.fetchall()
    
    if step_duplicates:
        print("\nFound duplicate sequence steps:")
        for dup in step_duplicates:
            print(f"\nSequence Step ID: {dup['sequence_stepid']}")
            print(f"Count: {dup['count']}")
            print(f"Message IDs: {dup['message_ids']}")
            print(f"Sent times: {dup['sent_times']}")
            print(f"Statuses: {dup['statuses']}")
    else:
        print("No duplicate sequence steps found")
        
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
