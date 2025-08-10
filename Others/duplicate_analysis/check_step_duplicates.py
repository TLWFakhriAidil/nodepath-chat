import pymysql

connection = pymysql.connect(
    host='159.89.198.71',
    user='admin_aqil',
    password='admin_aqil',
    database='admin_railway',
    charset='utf8mb4'
)

cursor = connection.cursor()

try:
    # Check for duplicate messages for the same sequence step
    print("=== CHECKING FOR DUPLICATE SEQUENCE STEPS FOR 60179075761 ===")
    
    query = """
    SELECT 
        sequence_stepid,
        COUNT(*) as msg_count,
        MIN(DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s')) as first_created,
        MAX(DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s')) as last_created,
        GROUP_CONCAT(id) as message_ids,
        GROUP_CONCAT(status) as statuses,
        GROUP_CONCAT(DATE_FORMAT(sent_at, '%H:%i:%s')) as sent_times
    FROM broadcast_messages
    WHERE recipient_phone = '60179075761'
    AND sequence_stepid IS NOT NULL
    GROUP BY sequence_stepid
    HAVING COUNT(*) > 1
    ORDER BY msg_count DESC
    """
    
    cursor.execute(query)
    duplicates = cursor.fetchall()
    
    if duplicates:
        print(f"\nFound {len(duplicates)} sequence steps with DUPLICATES:")
        for row in duplicates:
            print(f"\n{'='*60}")
            print(f"Sequence Step ID: {row[0]}")
            print(f"Number of duplicate messages: {row[1]}")
            print(f"First created: {row[2]}")
            print(f"Last created: {row[3]}")
            print(f"Message IDs: {row[4]}")
            print(f"Statuses: {row[5]}")
            print(f"Sent times: {row[6]}")
            
            # Get more details about this step
            cursor.execute("""
                SELECT 
                    id,
                    DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') as created,
                    DATE_FORMAT(scheduled_at, '%Y-%m-%d %H:%i:%s') as scheduled,
                    DATE_FORMAT(sent_at, '%Y-%m-%d %H:%i:%s') as sent,
                    status,
                    device_id
                FROM broadcast_messages
                WHERE sequence_stepid = %s
                AND recipient_phone = '60179075761'
                ORDER BY created_at
            """, (row[0],))
            
            details = cursor.fetchall()
            print("\nDetailed timeline:")
            for detail in details:
                print(f"  {detail[1]} - Created (Status: {detail[4]}, Scheduled: {detail[2]}, Sent: {detail[3]})")
    else:
        print("\nNo duplicates found")
    
    # Check the specific step from the screenshot
    print("\n\n=== CHECKING SPECIFIC STEP (ec4dce1b-d073-4efe-8fba-7c9b23bca883) ===")
    
    cursor.execute("""
        SELECT 
            id,
            DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') as created,
            DATE_FORMAT(sent_at, '%Y-%m-%d %H:%i:%s') as sent,
            status
        FROM broadcast_messages
        WHERE sequence_stepid = 'ec4dce1b-d073-4efe-8fba-7c9b23bca883'
        AND recipient_phone = '60179075761'
        ORDER BY created_at
    """)
    
    specific = cursor.fetchall()
    if specific:
        print(f"Found {len(specific)} messages for this step:")
        for s in specific:
            print(f"  ID: {s[0]}, Created: {s[1]}, Sent: {s[2]}, Status: {s[3]}")

finally:
    cursor.close()
    connection.close()
