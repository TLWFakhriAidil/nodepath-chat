import mysql.connector
from datetime import datetime
import pandas as pd

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
    
    # Check messages SENT around 1:38 PM (not scheduled)
    print("\n=== MESSAGES SENT AROUND 1:38 PM ===")
    query1 = """
    SELECT 
        id,
        recipient_phone,
        LEFT(content, 100) as message_preview,
        status,
        DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') as created_at,
        DATE_FORMAT(sent_at, '%Y-%m-%d %H:%i:%s') as sent_at,
        DATE_FORMAT(scheduled_at, '%Y-%m-%d %H:%i:%s') as scheduled_at,
        device_id,
        campaign_id,
        sequence_id,
        sequence_stepid,
        processing_worker_id
    FROM broadcast_messages
    WHERE (recipient_phone = '+60128198574' 
        OR recipient_phone = '60128198574'
        OR recipient_phone = '0128198574'
        OR recipient_phone LIKE '%128198574%')
        AND DATE(sent_at) = '2025-08-10'
        AND TIME(sent_at) BETWEEN '13:35:00' AND '13:40:00'
    ORDER BY sent_at
    """
    
    cursor.execute(query1)
    results = cursor.fetchall()
    
    if results:
        print(f"\nFound {len(results)} messages sent between 1:35 PM and 1:40 PM:")
        for msg in results:
            print(f"\nID: {msg['id']}")
            print(f"Phone: {msg['recipient_phone']}")
            print(f"Message: {msg['message_preview']}")
            print(f"Sent at: {msg['sent_at']}")
            print(f"Status: {msg['status']}")
            print(f"Device: {msg['device_id']}")
            print(f"Sequence: {msg['sequence_id']}")
            print(f"Step ID: {msg['sequence_stepid']}")
    else:
        print("No messages sent in this time range")
    
    # Check ALL messages for this phone number today (sent or created)
    print("\n\n=== ALL MESSAGES FOR THIS PHONE TODAY ===")
    query2 = """
    SELECT 
        id,
        LEFT(content, 80) as message_preview,
        status,
        DATE_FORMAT(created_at, '%H:%i:%s') as created_time,
        DATE_FORMAT(sent_at, '%H:%i:%s') as sent_time,
        DATE_FORMAT(scheduled_at, '%H:%i:%s') as scheduled_time,
        device_id,
        sequence_stepid
    FROM broadcast_messages
    WHERE (recipient_phone = '+60128198574' 
        OR recipient_phone = '60128198574'
        OR recipient_phone = '0128198574'
        OR recipient_phone LIKE '%128198574%')
        AND (DATE(created_at) = '2025-08-10' 
            OR DATE(sent_at) = '2025-08-10'
            OR DATE(scheduled_at) = '2025-08-10')
    ORDER BY COALESCE(sent_at, scheduled_at, created_at) DESC
    """
    
    cursor.execute(query2)
    all_msgs = cursor.fetchall()
    
    if all_msgs:
        df = pd.DataFrame(all_msgs)
        print(f"\nTotal messages found: {len(all_msgs)}")
        print(df.to_string(index=False))
    
    # Check for messages with same content
    print("\n\n=== CHECKING FOR DUPLICATE CONTENT ===")
    query3 = """
    SELECT 
        content,
        COUNT(*) as count,
        GROUP_CONCAT(id) as message_ids,
        GROUP_CONCAT(DATE_FORMAT(sent_at, '%H:%i:%s')) as sent_times,
        GROUP_CONCAT(status) as statuses
    FROM broadcast_messages
    WHERE (recipient_phone = '+60128198574' 
        OR recipient_phone = '60128198574'
        OR recipient_phone = '0128198574'
        OR recipient_phone LIKE '%128198574%')
        AND (DATE(created_at) = '2025-08-10' 
            OR DATE(sent_at) = '2025-08-10')
    GROUP BY content
    HAVING COUNT(*) > 1
    """
    
    cursor.execute(query3)
    duplicates = cursor.fetchall()
    
    if duplicates:
        print("\nFound duplicate content:")
        for dup in duplicates:
            print(f"\nContent: {dup['content'][:100]}...")
            print(f"Count: {dup['count']}")
            print(f"Message IDs: {dup['message_ids']}")
            print(f"Sent times: {dup['sent_times']}")
            print(f"Statuses: {dup['statuses']}")
    else:
        print("No duplicate content found")
        
except Exception as e:
    print(f"Error: {e}")
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
