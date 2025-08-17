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
    
    print("=== UPDATING SCHEDULED TIME TO NOW ===\n")
    
    # Update scheduled_at to current time
    update_query = """
    UPDATE broadcast_messages 
    SET scheduled_at = NOW(),
        updated_at = NOW()
    WHERE id = '1ef22b72-df41-45c1-9c69-be5925a38080'
    """
    
    cursor.execute(update_query)
    conn.commit()
    
    print("✅ Updated scheduled_at to NOW")
    
    # Verify the update
    verify_query = """
    SELECT 
        id,
        recipient_phone,
        status,
        DATE_FORMAT(scheduled_at, '%Y-%m-%d %H:%i:%s') as scheduled_at_utc,
        DATE_FORMAT(DATE_ADD(scheduled_at, INTERVAL 8 HOUR), '%Y-%m-%d %H:%i:%s') as scheduled_at_myt,
        device_id,
        processing_worker_id
    FROM broadcast_messages
    WHERE id = '1ef22b72-df41-45c1-9c69-be5925a38080'
    """
    
    cursor.execute(verify_query)
    result = cursor.fetchone()
    
    if result:
        print(f"\nUpdated Message Details:")
        print(f"ID: {result[0]}")
        print(f"Phone: {result[1]}")
        print(f"Status: {result[2]}")
        print(f"Scheduled (UTC): {result[3]}")
        print(f"Scheduled (MYT): {result[4]}")
        print(f"Device: {result[5]}")
        print(f"Worker ID: {result[6] or 'NULL'}")
        
        print("\n✅ Message is now within the 10-minute window!")
        print("It should be picked up by the processor within 5 seconds.")
        
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
