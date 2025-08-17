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
    
    print("=== CHECKING WHICH DEVICE HAS PENDING MESSAGES ===\n")
    
    # This query matches what GetDevicesWithPendingMessages uses
    query = """
    SELECT DISTINCT device_id, COUNT(*) as pending_count
    FROM broadcast_messages
    WHERE status = 'pending'
        AND scheduled_at IS NOT NULL
        AND scheduled_at <= DATE_ADD(NOW(), INTERVAL 8 HOUR)
    GROUP BY device_id
    ORDER BY pending_count DESC
    """
    
    cursor.execute(query)
    devices = cursor.fetchall()
    
    print(f"Found {len(devices)} devices with pending messages:\n")
    
    for device in devices:
        # Get device name
        cursor.execute("SELECT device_name, status, platform FROM user_devices WHERE id = %s", (device['device_id'],))
        device_info = cursor.fetchone()
        
        if device_info:
            print(f"Device: {device['device_id']}")
            print(f"Name: {device_info['device_name']}")
            print(f"Status: {device_info['status']}")
            print(f"Platform: {device_info['platform'] or 'WhatsApp Web'}")
            print(f"Pending: {device['pending_count']} messages")
            
            if device['device_id'] == '8badb299-f1d1-493a-bddf-84cbaba1273b':
                print("👆 THIS IS YOUR TEST MESSAGE DEVICE!")
        print("-" * 50)
    
    # Check why messages aren't being processed
    print("\n=== CHECKING WHY NOT PROCESSING ===\n")
    
    # Check if there are messages in the time window
    window_query = """
    SELECT COUNT(*) as in_window
    FROM broadcast_messages
    WHERE status = 'pending'
        AND scheduled_at IS NOT NULL
        AND scheduled_at <= DATE_ADD(NOW(), INTERVAL 8 HOUR)
        AND scheduled_at >= DATE_ADD(DATE_SUB(NOW(), INTERVAL 10 MINUTE), INTERVAL 8 HOUR)
    """
    
    cursor.execute(window_query)
    window = cursor.fetchone()
    
    print(f"Messages in 10-minute window: {window['in_window']}")
    
    if window['in_window'] == 0:
        print("❌ No messages in the time window!")
        print("The processor is finding devices but messages are too old.")
    else:
        print("✅ Messages are in the window")
        print("❌ But processor is NOT picking them up")
        print("This confirms OLD CODE is running!")
        
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
