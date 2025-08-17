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
    
    print("=== CHECKING AUGUST 11 PENDING MESSAGES ===\n")
    
    # Check all pending messages scheduled for Aug 11
    query = """
    SELECT 
        COUNT(*) as total_pending,
        COUNT(DISTINCT device_id) as devices_affected,
        MIN(DATE_FORMAT(scheduled_at, '%Y-%m-%d %H:%i:%s')) as earliest_scheduled,
        MAX(DATE_FORMAT(scheduled_at, '%Y-%m-%d %H:%i:%s')) as latest_scheduled,
        SUM(CASE WHEN processing_worker_id IS NOT NULL THEN 1 ELSE 0 END) as with_worker_id
    FROM broadcast_messages
    WHERE status = 'pending'
        AND DATE(scheduled_at) = '2025-08-11'
    """
    
    cursor.execute(query)
    summary = cursor.fetchone()
    
    print(f"Total pending for Aug 11: {summary['total_pending']}")
    print(f"Devices affected: {summary['devices_affected']}")
    print(f"Earliest scheduled: {summary['earliest_scheduled']}")
    print(f"Latest scheduled: {summary['latest_scheduled']}")
    print(f"Messages with worker ID: {summary['with_worker_id']}")
    
    # Check the time window issue
    print("\n=== TIME WINDOW ANALYSIS ===")
    
    window_query = """
    SELECT 
        COUNT(*) as total,
        SUM(CASE 
            WHEN scheduled_at >= DATE_ADD(DATE_SUB(NOW(), INTERVAL 10 MINUTE), INTERVAL 8 HOUR) 
            THEN 1 ELSE 0 
        END) as in_window,
        SUM(CASE 
            WHEN scheduled_at < DATE_ADD(DATE_SUB(NOW(), INTERVAL 10 MINUTE), INTERVAL 8 HOUR) 
            THEN 1 ELSE 0 
        END) as outside_window
    FROM broadcast_messages
    WHERE status = 'pending'
        AND DATE(scheduled_at) = '2025-08-11'
    """
    
    cursor.execute(window_query)
    window = cursor.fetchone()
    
    print(f"Total pending: {window['total']}")
    print(f"In 10-minute window: {window['in_window']}")
    print(f"Outside window (too old): {window['outside_window']}")
    
    # Check by device status
    print("\n=== DEVICE STATUS BREAKDOWN ===")
    
    device_query = """
    SELECT 
        d.device_name,
        d.status as device_status,
        d.platform,
        COUNT(bm.id) as pending_count,
        MIN(DATE_FORMAT(bm.scheduled_at, '%H:%i:%s')) as earliest_time,
        MAX(DATE_FORMAT(bm.scheduled_at, '%H:%i:%s')) as latest_time
    FROM broadcast_messages bm
    JOIN user_devices d ON bm.device_id = d.id
    WHERE bm.status = 'pending'
        AND DATE(bm.scheduled_at) = '2025-08-11'
    GROUP BY d.id, d.device_name, d.status, d.platform
    ORDER BY pending_count DESC
    LIMIT 10
    """
    
    cursor.execute(device_query)
    devices = cursor.fetchall()
    
    if devices:
        print("\nTop devices with pending messages:")
        for device in devices:
            print(f"\nDevice: {device['device_name']}")
            print(f"Status: {device['device_status']}")
            print(f"Platform: {device['platform'] or 'WhatsApp Web'}")
            print(f"Pending: {device['pending_count']} messages")
            print(f"Time range: {device['earliest_time']} - {device['latest_time']}")
    
    # Check recent processing activity
    print("\n\n=== RECENT PROCESSING ACTIVITY ===")
    
    activity_query = """
    SELECT 
        DATE_FORMAT(sent_at, '%Y-%m-%d %H:00') as hour,
        COUNT(*) as messages_sent
    FROM broadcast_messages
    WHERE sent_at >= DATE_SUB(NOW(), INTERVAL 6 HOUR)
    GROUP BY DATE_FORMAT(sent_at, '%Y-%m-%d %H:00')
    ORDER BY hour DESC
    """
    
    cursor.execute(activity_query)
    activity = cursor.fetchall()
    
    if activity:
        print("\nMessages sent by hour:")
        for hour in activity:
            print(f"{hour['hour']}: {hour['messages_sent']} messages")
    else:
        print("No messages sent in last 6 hours!")
    
    # Check if processor is finding these messages
    print("\n\n=== PROCESSOR STATUS CHECK ===")
    
    # Check messages that should be processed NOW
    now_query = """
    SELECT COUNT(*) as should_process
    FROM broadcast_messages
    WHERE status = 'pending'
        AND scheduled_at <= DATE_ADD(NOW(), INTERVAL 8 HOUR)
        AND scheduled_at >= DATE_ADD(DATE_SUB(NOW(), INTERVAL 10 MINUTE), INTERVAL 8 HOUR)
        AND processing_worker_id IS NULL
    """
    
    cursor.execute(now_query)
    should_process = cursor.fetchone()
    
    print(f"Messages that should be processed NOW: {should_process['should_process']}")
    
    if should_process['should_process'] > 0:
        print("\n❌ PROCESSOR IS NOT PICKING UP ELIGIBLE MESSAGES!")
        print("Possible causes:")
        print("1. Processor is stopped")
        print("2. Old code is running (no worker ID assignment)")
        print("3. Database connection issues")
    
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
