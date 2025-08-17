import pymysql
from datetime import datetime

connection = pymysql.connect(
    host='159.89.198.71',
    user='admin_aqil',
    password='admin_aqil',
    database='admin_railway',
    charset='utf8mb4'
)

cursor = connection.cursor()

try:
    # Check messages sent around 1:23-1:25 PM (13:23-13:25)
    print("=== CHECKING MESSAGES SENT BETWEEN 1:23 PM - 1:25 PM ===")
    
    query = """
    SELECT 
        id,
        recipient_phone,
        sequence_stepid,
        DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') as created,
        DATE_FORMAT(sent_at, '%Y-%m-%d %H:%i:%s') as sent,
        status,
        device_id,
        error_message
    FROM broadcast_messages
    WHERE recipient_phone = '60179075761'
    AND DATE(sent_at) = '2025-08-06'
    AND TIME(sent_at) BETWEEN '13:20:00' AND '13:30:00'
    ORDER BY sent_at
    """
    
    cursor.execute(query)
    results = cursor.fetchall()
    
    print(f"\nFound {len(results)} messages sent in that time window")
    
    for row in results:
        print(f"\n--- Message ---")
        print(f"ID: {row[0]}")
        print(f"Phone: {row[1]}")
        print(f"Step ID: {row[2]}")
        print(f"Created: {row[3]}")
        print(f"Sent: {row[4]}")
        print(f"Status: {row[5]}")
        print(f"Device: {row[6]}")
        print(f"Error: {row[7]}")
    
    # Check the worker logs or broadcast history
    print("\n\n=== CHECKING ALL SENT MESSAGES FOR THIS PHONE TODAY ===")
    
    cursor.execute("""
        SELECT 
            DATE_FORMAT(sent_at, '%H:%i:%s') as sent_time,
            id,
            sequence_stepid,
            status
        FROM broadcast_messages
        WHERE recipient_phone = '60179075761'
        AND DATE(sent_at) = '2025-08-06'
        AND sent_at IS NOT NULL
        ORDER BY sent_at
    """)
    
    sent_today = cursor.fetchall()
    
    print(f"\nTotal messages sent today: {len(sent_today)}")
    for msg in sent_today:
        print(f"  {msg[0]} - ID: {msg[1]}, Step: {msg[2]}, Status: {msg[3]}")

finally:
    cursor.close()
    connection.close()
