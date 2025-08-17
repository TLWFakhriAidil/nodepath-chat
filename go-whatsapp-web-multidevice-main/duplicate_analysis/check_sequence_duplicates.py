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
    # First, let's find ALL messages for this phone number today
    print("=== CHECKING ALL MESSAGES FOR +60179075761 ===")
    
    query1 = """
    SELECT 
        id,
        recipient_phone,
        device_id,
        campaign_id,
        sequence_id,
        sequence_stepid,
        status,
        DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') as created_at,
        DATE_FORMAT(updated_at, '%Y-%m-%d %H:%i:%s') as updated_at,
        DATE_FORMAT(sent_at, '%Y-%m-%d %H:%i:%s') as sent_at,
        DATE_FORMAT(scheduled_at, '%Y-%m-%d %H:%i:%s') as scheduled_at,
        LEFT(content, 100) as content_preview
    FROM broadcast_messages
    WHERE recipient_phone = '+60179075761'
    AND DATE(created_at) >= '2025-08-06'
    ORDER BY created_at DESC
    LIMIT 20
    """
    
    cursor.execute(query1)
    results = cursor.fetchall()
    
    print(f"\nFound {len(results)} messages total")
    
    for row in results:
        print(f"\n--- Message ---")
        print(f"ID: {row[0]}")
        print(f"Phone: {row[1]}")
        print(f"Device: {row[2]}")
        print(f"Campaign: {row[3]}")
        print(f"Sequence: {row[4]}")
        print(f"Step ID: {row[5]}")
        print(f"Status: {row[6]}")
        print(f"Created: {row[7]}")
        print(f"Updated: {row[8]}")
        print(f"Sent: {row[9]}")
        print(f"Scheduled: {row[10]}")
        print(f"Content: {row[11]}...")
    
    # Now check for duplicates by sequence_stepid
    print("\n\n=== CHECKING FOR DUPLICATE SEQUENCE STEPS ===")
    
    query2 = """
    SELECT 
        sequence_stepid,
        COUNT(*) as count,
        GROUP_CONCAT(id) as message_ids,
        GROUP_CONCAT(status) as statuses,
        GROUP_CONCAT(DATE_FORMAT(created_at, '%H:%i:%s')) as created_times
    FROM broadcast_messages
    WHERE recipient_phone = '+60179075761'
    AND sequence_stepid IS NOT NULL
    AND DATE(created_at) >= '2025-08-06'
    GROUP BY sequence_stepid
    HAVING COUNT(*) > 1
    """
    
    cursor.execute(query2)
    duplicates = cursor.fetchall()
    
    if duplicates:
        print(f"\nFound {len(duplicates)} sequence steps with duplicates:")
        for dup in duplicates:
            print(f"\nSequence Step ID: {dup[0]}")
            print(f"Count: {dup[1]} messages")
            print(f"Message IDs: {dup[2]}")
            print(f"Statuses: {dup[3]}")
            print(f"Created times: {dup[4]}")
    else:
        print("\nNo duplicate sequence steps found")
    
    # Check sequence_contacts status
    print("\n\n=== CHECKING SEQUENCE CONTACTS STATUS ===")
    
    query3 = """
    SELECT 
        sc.id,
        sc.sequence_id,
        sc.contact_phone,
        sc.current_step,
        sc.status,
        DATE_FORMAT(sc.enrolled_at, '%Y-%m-%d %H:%i:%s') as enrolled_at,
        DATE_FORMAT(sc.last_message_at, '%Y-%m-%d %H:%i:%s') as last_message_at,
        s.name as sequence_name
    FROM sequence_contacts sc
    LEFT JOIN sequences s ON sc.sequence_id = s.id
    WHERE sc.contact_phone = '+60179075761'
    AND sc.status = 'active'
    """
    
    cursor.execute(query3)
    contacts = cursor.fetchall()
    
    print(f"\nFound {len(contacts)} active sequence enrollments:")
    for contact in contacts:
        print(f"\nSequence: {contact[7]}")
        print(f"Current Step: {contact[3]}")
        print(f"Status: {contact[4]}")
        print(f"Enrolled: {contact[5]}")
        print(f"Last Message: {contact[6]}")

finally:
    cursor.close()
    connection.close()
