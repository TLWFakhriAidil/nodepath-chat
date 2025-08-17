import mysql.connector
from datetime import datetime, timedelta
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
    
    print("=== SETTING MESSAGE TO BE ELIGIBLE NOW ===\n")
    
    # Set scheduled_at to 5 minutes ago
    update_query = """
    UPDATE broadcast_messages 
    SET scheduled_at = DATE_SUB(NOW(), INTERVAL 5 MINUTE),
        updated_at = NOW()
    WHERE id = '1ef22b72-df41-45c1-9c69-be5925a38080'
    """
    
    cursor.execute(update_query)
    conn.commit()
    
    print("✅ Updated scheduled time to 5 minutes ago")
    
    # Verify
    verify = """
    SELECT 
        status,
        DATE_FORMAT(scheduled_at, '%Y-%m-%d %H:%i:%s') as scheduled_at,
        DATE_FORMAT(NOW(), '%Y-%m-%d %H:%i:%s') as now_time,
        processing_worker_id
    FROM broadcast_messages
    WHERE id = '1ef22b72-df41-45c1-9c69-be5925a38080'
    """
    
    cursor.execute(verify)
    result = cursor.fetchone()
    
    print(f"\nStatus: {result[0]}")
    print(f"Scheduled: {result[1]} UTC")
    print(f"Current: {result[2]} UTC")
    print(f"Worker ID: {result[3] or 'NULL'}")
    
    print("\n✅ Message is now eligible for processing!")
    print("\n⚠️  BUT: The processor appears to be STOPPED!")
    print("- Last message sent: 2+ hours ago")
    print("- No worker IDs are being set")
    print("- Application needs to be restarted with the new code")
    
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
