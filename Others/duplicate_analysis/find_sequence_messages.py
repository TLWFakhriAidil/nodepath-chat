import pymysql
from datetime import datetime, timedelta

# MySQL connection
connection = pymysql.connect(
    host='159.89.198.71',
    user='admin_aqil',
    password='admin_aqil',
    database='admin_railway',
    charset='utf8mb4'
)

cursor = connection.cursor()

try:
    # Check the phone number format - might need to check without +
    print("=== CHECKING MESSAGES WITH DIFFERENT PHONE FORMATS ===")
    
    phone_formats = ['+60179075761', '60179075761', '0179075761', '+6017-907-5761', '6017-907-5761']
    
    for phone in phone_formats:
        query = """
        SELECT COUNT(*) 
        FROM broadcast_messages 
        WHERE recipient_phone = %s 
        AND DATE(created_at) >= '2025-08-05'
        """
        cursor.execute(query, (phone,))
        count = cursor.fetchone()[0]
        if count > 0:
            print(f"Found {count} messages for phone format: {phone}")
    
    # Get all messages from today regardless of phone
    print("\n=== RECENT SEQUENCE MESSAGES (LAST 2 DAYS) ===")
    
    query2 = """
    SELECT 
        id,
        recipient_phone,
        device_id,
        sequence_id,
        sequence_stepid,
        status,
        DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') as created_at,
        DATE_FORMAT(sent_at, '%Y-%m-%d %H:%i:%s') as sent_at,
        LEFT(content, 80) as content_preview
    FROM broadcast_messages
    WHERE sequence_id IS NOT NULL
    AND DATE(created_at) >= DATE_SUB(CURDATE(), INTERVAL 2 DAY)
    AND (recipient_phone LIKE '%9075761%' OR recipient_phone LIKE '%907-5761%')
    ORDER BY created_at DESC
    LIMIT 30
    """
    
    cursor.execute(query2)
    results = cursor.fetchall()
    
    print(f"\nFound {len(results)} sequence messages")
    
    sequence_steps = {}
    for row in results:
        print(f"\n--- Message ---")
        print(f"ID: {row[0]}")
        print(f"Phone: {row[1]}")
        print(f"Device: {row[2]}")
        print(f"Sequence: {row[3]}")
        print(f"Step ID: {row[4]}")
        print(f"Status: {row[5]}")
        print(f"Created: {row[6]}")
        print(f"Sent: {row[7]}")
        print(f"Content: {row[8]}...")
        
        # Track duplicates
        key = f"{row[1]}_{row[4]}"  # phone_stepid
        if key not in sequence_steps:
            sequence_steps[key] = []
        sequence_steps[key].append({
            'id': row[0],
            'created': row[6],
            'sent': row[7],
            'status': row[5]
        })
    
    # Show duplicates
    print("\n\n=== DUPLICATE ANALYSIS ===")
    for key, messages in sequence_steps.items():
        if len(messages) > 1:
            phone, step_id = key.split('_')
            print(f"\nDUPLICATE FOUND!")
            print(f"Phone: {phone}")
            print(f"Step ID: {step_id}")
            print(f"Count: {len(messages)} messages")
            for msg in messages:
                print(f"  - ID: {msg['id']}, Created: {msg['created']}, Sent: {msg['sent']}, Status: {msg['status']}")

finally:
    cursor.close()
    connection.close()
