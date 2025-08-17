import mysql.connector
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
    
    print("=== CHECKING WORKER ID IMPLEMENTATION ===\n")
    
    # Check messages for 60128198574 with focus on worker fields
    query1 = """
    SELECT 
        id,
        sequence_stepid,
        status,
        processing_worker_id,
        processing_started_at,
        DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') as created_at,
        DATE_FORMAT(sent_at, '%Y-%m-%d %H:%i:%s') as sent_at,
        device_id
    FROM broadcast_messages
    WHERE recipient_phone LIKE '%128198574%'
    ORDER BY created_at DESC
    """
    
    cursor.execute(query1)
    results = cursor.fetchall()
    
    print(f"Messages for 60128198574 - Worker ID Status:\n")
    for msg in results:
        worker_status = "❌ NULL" if msg['processing_worker_id'] is None else f"✅ {msg['processing_worker_id']}"
        started_status = "❌ NULL" if msg['processing_started_at'] is None else f"✅ {msg['processing_started_at']}"
        
        print(f"ID: {msg['id'][:8]}...")
        print(f"Status: {msg['status']}")
        print(f"Worker ID: {worker_status}")
        print(f"Processing Started: {started_status}")
        print(f"Created: {msg['created_at']}")
        print(f"Sent: {msg['sent_at']}")
        print("-" * 50)
    
    # Check overall worker ID usage
    print("\n\n=== OVERALL WORKER ID USAGE ===\n")
    
    query2 = """
    SELECT 
        DATE(created_at) as date,
        COUNT(*) as total_messages,
        SUM(CASE WHEN processing_worker_id IS NOT NULL THEN 1 ELSE 0 END) as with_worker_id,
        SUM(CASE WHEN processing_worker_id IS NULL THEN 1 ELSE 0 END) as without_worker_id,
        ROUND(SUM(CASE WHEN processing_worker_id IS NOT NULL THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 2) as percentage_with_worker
    FROM broadcast_messages
    WHERE created_at >= DATE_SUB(NOW(), INTERVAL 3 DAY)
    GROUP BY DATE(created_at)
    ORDER BY date DESC
    """
    
    cursor.execute(query2)
    stats = cursor.fetchall()
    
    for stat in stats:
        status = "❌ NOT WORKING" if stat['percentage_with_worker'] == 0 else f"⚠️ PARTIAL ({stat['percentage_with_worker']}%)"
        print(f"Date: {stat['date']}")
        print(f"Total: {stat['total_messages']} messages")
        print(f"With Worker ID: {stat['with_worker_id']} ({stat['percentage_with_worker']}%)")
        print(f"Without Worker ID: {stat['without_worker_id']}")
        print(f"Status: {status}")
        print()
    
    # Check for race conditions - multiple messages being processed at same time
    print("\n=== CHECKING FOR RACE CONDITIONS ===\n")
    
    query3 = """
    SELECT 
        recipient_phone,
        COUNT(*) as message_count,
        GROUP_CONCAT(id) as message_ids,
        GROUP_CONCAT(status) as statuses,
        GROUP_CONCAT(processing_worker_id) as worker_ids,
        GROUP_CONCAT(DATE_FORMAT(processing_started_at, '%H:%i:%s')) as started_times,
        MIN(processing_started_at) as first_started,
        MAX(processing_started_at) as last_started,
        TIMESTAMPDIFF(SECOND, MIN(processing_started_at), MAX(processing_started_at)) as time_spread
    FROM broadcast_messages
    WHERE status IN ('sent', 'processing')
        AND created_at >= DATE_SUB(NOW(), INTERVAL 1 DAY)
    GROUP BY recipient_phone
    HAVING COUNT(*) > 1 AND time_spread < 60
    ORDER BY message_count DESC
    LIMIT 10
    """
    
    cursor.execute(query3)
    races = cursor.fetchall()
    
    if races:
        print("❌ FOUND POTENTIAL RACE CONDITIONS:\n")
        for race in races:
            print(f"Phone: {race['recipient_phone']}")
            print(f"Messages: {race['message_count']} within {race['time_spread']} seconds")
            print(f"Worker IDs: {race['worker_ids']}")
            print(f"Started times: {race['started_times']}")
            print("-" * 50)
    else:
        print("No obvious race conditions found")
        
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
