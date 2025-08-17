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
    cursor = conn.cursor()
    
    print("=== FIXING TEST MESSAGE ===\n")
    
    # Update scheduled_at to NOW so it gets picked up immediately
    update_query = """
    UPDATE broadcast_messages 
    SET scheduled_at = NOW()
    WHERE id = '1ef22b72-df41-45c1-9c69-be5925a38080'
    """
    
    cursor.execute(update_query)
    conn.commit()
    
    print("✅ Updated scheduled_at to NOW")
    
    # Also, let's find a WhatsApp Web device that's online
    print("\n=== FINDING ONLINE WHATSAPP WEB DEVICE ===\n")
    
    query = """
    SELECT 
        id,
        device_name,
        status,
        platform
    FROM user_devices
    WHERE status IN ('connected', 'online')
        AND (platform IS NULL OR platform = '')
    LIMIT 5
    """
    
    cursor.execute(query)
    devices = cursor.fetchall()
    
    if devices:
        print("Found online WhatsApp Web devices:")
        for device in devices:
            print(f"- {device[0]}: {device[1]} (Status: {device[2]})")
        
        # Update the test message to use the first online device
        first_device = devices[0][0]
        update_device = """
        UPDATE broadcast_messages 
        SET device_id = %s
        WHERE id = '1ef22b72-df41-45c1-9c69-be5925a38080'
        """
        
        cursor.execute(update_device, (first_device,))
        conn.commit()
        
        print(f"\n✅ Updated device to: {first_device} ({devices[0][1]})")
    else:
        print("❌ No online WhatsApp Web devices found!")
        
        # Check platform devices
        query2 = """
        SELECT 
            id,
            device_name,
            status,
            platform
        FROM user_devices
        WHERE status IN ('connected', 'online')
            AND platform IS NOT NULL
        LIMIT 5
        """
        
        cursor.execute(query2)
        platform_devices = cursor.fetchall()
        
        if platform_devices:
            print("\nFound online platform devices:")
            for device in platform_devices:
                print(f"- {device[0]}: {device[1]} (Platform: {device[3]}, Status: {device[2]})")
    
    # Verify the update
    print("\n=== VERIFYING UPDATE ===\n")
    
    verify_query = """
    SELECT 
        id,
        recipient_phone,
        status,
        device_id,
        DATE_FORMAT(scheduled_at, '%Y-%m-%d %H:%i:%s') as scheduled_at,
        processing_worker_id
    FROM broadcast_messages
    WHERE id = '1ef22b72-df41-45c1-9c69-be5925a38080'
    """
    
    cursor.execute(verify_query)
    result = cursor.fetchone()
    
    if result:
        print(f"ID: {result[0]}")
        print(f"Phone: {result[1]}")
        print(f"Status: {result[2]}")
        print(f"Device: {result[3]}")
        print(f"Scheduled: {result[4]}")
        print(f"Worker ID: {result[5] or 'NULL'}")
        
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
