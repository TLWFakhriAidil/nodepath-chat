import mysql.connector
from datetime import datetime
import pandas as pd

# Database connection from your .env file
config = {
    'host': '159.89.198.71',
    'port': 3306,
    'database': 'admin_railway',
    'user': 'admin_aqil',
    'password': 'admin_aqil'
}

try:
    # Connect to MySQL
    conn = mysql.connector.connect(**config)
    cursor = conn.cursor(dictionary=True)
    print("Connected to MySQL database successfully!")
    
    # Query 1: Check all messages for this phone number today
    print("\n=== ALL MESSAGES FOR +60128198574 TODAY (2025-08-10) ===")
    query1 = """
    SELECT 
        id,
        recipient_phone,
        LEFT(content, 80) as message_preview,
        status,
        DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') as created_at,
        DATE_FORMAT(sent_at, '%Y-%m-%d %H:%i:%s') as sent_at,
        DATE_FORMAT(scheduled_at, '%Y-%m-%d %H:%i:%s') as scheduled_at,
        device_id,
        campaign_id,
        sequence_id,
        processing_worker_id
    FROM broadcast_messages
    WHERE (recipient_phone = '+60128198574' 
        OR recipient_phone = '60128198574'
        OR recipient_phone = '0128198574'
        OR recipient_phone LIKE '%128198574%')
        AND DATE(scheduled_at) = '2025-08-10'
    ORDER BY scheduled_at DESC
    """
    
    cursor.execute(query1)
    results = cursor.fetchall()
    
    if results:
        df = pd.DataFrame(results)
        print(f"\nFound {len(results)} messages scheduled for today:")
        print(df.to_string(index=False))
    else:
        print("No messages found for this phone number scheduled today")
    
    # Query 2: Check for duplicates by content
    print("\n\n=== DUPLICATE MESSAGES (SAME CONTENT) ===")
    query2 = """
    SELECT 
        recipient_phone,
        LEFT(content, 80) as message_preview,
        COUNT(*) as duplicate_count,
        DATE_FORMAT(MIN(scheduled_at), '%Y-%m-%d %H:%i:%s') as first_scheduled,
        DATE_FORMAT(MAX(scheduled_at), '%Y-%m-%d %H:%i:%s') as last_scheduled,
        GROUP_CONCAT(status) as all_statuses,
        GROUP_CONCAT(DISTINCT device_id) as devices_used
    FROM broadcast_messages
    WHERE (recipient_phone = '+60128198574' 
        OR recipient_phone = '60128198574'
        OR recipient_phone = '0128198574'
        OR recipient_phone LIKE '%128198574%')
        AND DATE(scheduled_at) = '2025-08-10'
    GROUP BY recipient_phone, content
    HAVING COUNT(*) > 1
    """
    
    cursor.execute(query2)
    duplicates = cursor.fetchall()
    
    if duplicates:
        df_dup = pd.DataFrame(duplicates)
        print(f"\nFound duplicate messages:")
        print(df_dup.to_string(index=False))
    else:
        print("No duplicate messages found")
    
    # Query 3: Check messages around 1:38 PM
    print("\n\n=== MESSAGES AROUND 1:38 PM (13:38) ===")
    query3 = """
    SELECT 
        id,
        recipient_phone,
        LEFT(content, 50) as message_preview,
        status,
        DATE_FORMAT(created_at, '%H:%i:%s') as created_time,
        DATE_FORMAT(sent_at, '%H:%i:%s') as sent_time,
        DATE_FORMAT(scheduled_at, '%H:%i:%s') as scheduled_time,
        processing_worker_id,
        device_id,
        campaign_id,
        sequence_id
    FROM broadcast_messages
    WHERE (recipient_phone = '+60128198574' 
        OR recipient_phone = '60128198574'
        OR recipient_phone = '0128198574'
        OR recipient_phone LIKE '%128198574%')
        AND DATE(scheduled_at) = '2025-08-10'
        AND HOUR(scheduled_at) BETWEEN 13 AND 14
    ORDER BY scheduled_at
    """
    
    cursor.execute(query3)
    time_results = cursor.fetchall()
    
    if time_results:
        df_time = pd.DataFrame(time_results)
        print(f"\nFound {len(time_results)} messages scheduled between 1:00 PM and 2:00 PM:")
        print(df_time.to_string(index=False))
    else:
        print("No messages found in this time range")
    
    # Query 4: Check if unique constraints exist
    print("\n\n=== CHECKING UNIQUE CONSTRAINTS ===")
    query4 = """
    SELECT 
        CONSTRAINT_NAME,
        COLUMN_NAME
    FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
    WHERE TABLE_NAME = 'broadcast_messages'
        AND TABLE_SCHEMA = 'admin_railway'
        AND CONSTRAINT_NAME LIKE 'unique_%'
    """
    
    cursor.execute(query4)
    constraints = cursor.fetchall()
    
    if constraints:
        print("Unique constraints found:")
        for constraint in constraints:
            print(f"- {constraint['CONSTRAINT_NAME']}: {constraint['COLUMN_NAME']}")
    else:
        print("No unique constraints found on broadcast_messages table")
        
except mysql.connector.Error as err:
    print(f"MySQL Error: {err}")
except Exception as e:
    print(f"Error: {e}")
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
    print("\n\nDatabase connection closed.")
