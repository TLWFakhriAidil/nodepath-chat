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
    # Check the specific message that was resent
    message_id = '31ddccd3-e0d5-4abe-b015-15106daaeb64'
    
    print(f"=== CHECKING MESSAGE {message_id} ===")
    
    # Get full history of this message
    query = """
    SELECT 
        id,
        status,
        DATE_FORMAT(created_at, '%%Y-%%m-%%d %%H:%%i:%%s') as created_at,
        DATE_FORMAT(updated_at, '%%Y-%%m-%%d %%H:%%i:%%s') as updated_at,
        DATE_FORMAT(sent_at, '%%Y-%%m-%%d %%H:%%i:%%s') as sent_at,
        DATE_FORMAT(scheduled_at, '%%Y-%%m-%%d %%H:%%i:%%s') as scheduled_at,
        error_message,
        device_id,
        sequence_stepid
    FROM broadcast_messages
    WHERE id = %s
    """
    
    cursor.execute(query, (message_id,))
    result = cursor.fetchone()
    
    if result:
        print(f"\nMessage Details:")
        print(f"ID: {result[0]}")
        print(f"Status: {result[1]}")
        print(f"Created: {result[2]}")
        print(f"Updated: {result[3]}")
        print(f"Sent: {result[4]}")
        print(f"Scheduled: {result[5]}")
        print(f"Error: {result[6]}")
        print(f"Device: {result[7]}")
        print(f"Step ID: {result[8]}")
    
    # Check if there are any audit logs or history tables
    print("\n\n=== CHECKING FOR MESSAGE HISTORY/AUDIT ===")
    
    # Check if message appears in any logs
    cursor.execute("SHOW TABLES LIKE '%log%'")
    log_tables = cursor.fetchall()
    print(f"Found {len(log_tables)} log tables: {log_tables}")
    
    # Check scheduled_at time vs sent time
    print("\n\n=== TIME ANALYSIS ===")
    if result:
        created = datetime.strptime(result[2], '%Y-%m-%d %H:%M:%S')
        if result[4]:  # sent_at exists
            sent = datetime.strptime(result[4], '%Y-%m-%d %H:%M:%S')
            diff = sent - created
            print(f"Time between created and sent: {diff}")
            print(f"Sent at 5:23 AM but you received at 1:23 PM")
            print(f"That's an 8-hour difference!")
        
        if result[5]:  # scheduled_at exists
            scheduled = datetime.strptime(result[5], '%Y-%m-%d %H:%M:%S')
            print(f"Scheduled time: {scheduled}")
    
    # Check for any timezone issues
    cursor.execute("SELECT NOW(), UTC_TIMESTAMP()")
    server_time = cursor.fetchone()
    print(f"\nServer time: {server_time[0]}")
    print(f"UTC time: {server_time[1]}")
    
    # Check if there are multiple devices sending the same message
    print("\n\n=== CHECKING FOR DEVICE ISSUES ===")
    cursor.execute("""
        SELECT 
            device_id,
            COUNT(*) as count
        FROM broadcast_messages
        WHERE sequence_stepid = %s
        AND recipient_phone = '60179075761'
        GROUP BY device_id
    """, (result[8],))
    
    devices = cursor.fetchall()
    print(f"Messages per device for this step:")
    for dev in devices:
        print(f"  Device {dev[0]}: {dev[1]} messages")

finally:
    cursor.close()
    connection.close()
