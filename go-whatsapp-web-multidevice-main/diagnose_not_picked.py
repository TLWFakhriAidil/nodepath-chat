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
    
    print("=== COMPREHENSIVE CHECK - WHY MESSAGE NOT PICKED UP ===\n")
    
    # 1. Check the exact conditions the processor looks for
    print("1. CHECKING PROCESSOR CONDITIONS:")
    
    # This is the exact query from GetPendingMessagesAndLock
    query1 = """
    SELECT COUNT(*) as eligible_count
    FROM broadcast_messages 
    WHERE device_id = '8badb299-f1d1-493a-bddf-84cbaba1273b'
    AND status = 'pending'
    AND processing_worker_id IS NULL
    AND scheduled_at IS NOT NULL
    AND scheduled_at <= DATE_ADD(NOW(), INTERVAL 8 HOUR)
    AND scheduled_at >= DATE_ADD(DATE_SUB(NOW(), INTERVAL 10 MINUTE), INTERVAL 8 HOUR)
    """
    
    cursor.execute(query1)
    eligible = cursor.fetchone()
    print(f"Messages eligible for processing: {eligible['eligible_count']}")
    
    # 2. Check if ANY messages are being processed
    print("\n2. CHECKING OVERALL SYSTEM ACTIVITY:")
    
    query2 = """
    SELECT 
        COUNT(*) as total_sent_today,
        MAX(sent_at) as last_sent_time,
        MIN(sent_at) as first_sent_time
    FROM broadcast_messages
    WHERE DATE(sent_at) = CURDATE()
    """
    
    cursor.execute(query2)
    activity = cursor.fetchone()
    print(f"Messages sent today: {activity['total_sent_today']}")
    print(f"First sent: {activity['first_sent_time']}")
    print(f"Last sent: {activity['last_sent_time']}")
    
    # 3. Check if processor is running AT ALL
    print("\n3. CHECKING IF PROCESSOR IS ACTIVE:")
    
    query3 = """
    SELECT 
        status,
        COUNT(*) as count,
        MAX(updated_at) as last_update
    FROM broadcast_messages
    WHERE updated_at >= DATE_SUB(NOW(), INTERVAL 5 MINUTE)
    GROUP BY status
    """
    
    cursor.execute(query3)
    recent_updates = cursor.fetchall()
    
    if recent_updates:
        print("Recent status changes (last 5 minutes):")
        for update in recent_updates:
            print(f"- {update['status']}: {update['count']} messages, last: {update['last_update']}")
    else:
        print("❌ NO STATUS CHANGES IN LAST 5 MINUTES - PROCESSOR MIGHT BE STOPPED!")
    
    # 4. Check if device has pending messages
    print("\n4. DEVICE MESSAGE BACKLOG:")
    
    query4 = """
    SELECT 
        d.device_name,
        d.status as device_status,
        d.platform,
        COUNT(CASE WHEN bm.status = 'pending' THEN 1 END) as pending_count,
        COUNT(CASE WHEN bm.status = 'sent' AND DATE(bm.sent_at) = CURDATE() THEN 1 END) as sent_today,
        MIN(CASE WHEN bm.status = 'pending' THEN bm.created_at END) as oldest_pending
    FROM user_devices d
    LEFT JOIN broadcast_messages bm ON d.id = bm.device_id
    WHERE d.id = '8badb299-f1d1-493a-bddf-84cbaba1273b'
    GROUP BY d.id, d.device_name, d.status, d.platform
    """
    
    cursor.execute(query4)
    device_info = cursor.fetchone()
    
    if device_info:
        print(f"Device: {device_info['device_name']}")
        print(f"Status: {device_info['device_status']}")
        print(f"Platform: {device_info['platform']}")
        print(f"Pending messages: {device_info['pending_count']}")
        print(f"Sent today: {device_info['sent_today']}")
        print(f"Oldest pending: {device_info['oldest_pending']}")
    
    # 5. Check if there's a system-wide issue
    print("\n5. SYSTEM-WIDE CHECK:")
    
    query5 = """
    SELECT 
        COUNT(DISTINCT device_id) as devices_with_pending,
        COUNT(*) as total_pending,
        SUM(CASE WHEN processing_worker_id IS NOT NULL THEN 1 ELSE 0 END) as with_worker_id
    FROM broadcast_messages
    WHERE status = 'pending'
    """
    
    cursor.execute(query5)
    system = cursor.fetchone()
    
    print(f"Devices with pending messages: {system['devices_with_pending']}")
    print(f"Total pending messages: {system['total_pending']}")
    print(f"Messages with worker ID: {system['with_worker_id']}")
    
    if system['with_worker_id'] == 0 and system['total_pending'] > 0:
        print("\n❌ CRITICAL: NO MESSAGES HAVE WORKER IDs!")
        print("This confirms the OLD CODE is running or the processor is STOPPED!")
    
    # 6. Final diagnosis
    print("\n=== DIAGNOSIS ===")
    
    if eligible['eligible_count'] > 0 and activity['total_sent_today'] == 0:
        print("❌ PROCESSOR IS NOT RUNNING - No messages sent today despite eligible messages")
    elif eligible['eligible_count'] > 0 and not recent_updates:
        print("❌ PROCESSOR IS STUCK OR STOPPED - No recent activity")
    elif device_info['device_status'] != 'online':
        print("❌ DEVICE IS OFFLINE - Cannot send messages")
    elif eligible['eligible_count'] == 0:
        print("❌ MESSAGE DOESN'T MEET CONDITIONS - Check scheduled time")
    else:
        print("❌ UNKNOWN ISSUE - Need to check application logs")
        
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
