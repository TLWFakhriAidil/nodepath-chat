import mysql.connector

# Database connection
db_config = {
    'host': '159.89.198.71',
    'port': 3306,
    'user': 'admin_aqil',
    'password': 'admin_aqil',
    'database': 'admin_railway'
}

try:
    conn = mysql.connector.connect(**db_config)
    cursor = conn.cursor()
    
    print("Fixing NULL status in user_devices table...")
    print("=" * 60)
    
    # Check how many NULL values we have in user_devices
    cursor.execute("SELECT COUNT(*) FROM user_devices WHERE status IS NULL")
    null_count = cursor.fetchone()[0]
    print(f"Found {null_count} devices with NULL status")
    
    if null_count > 0:
        # Update NULL values to 'offline'
        cursor.execute("UPDATE user_devices SET status = 'offline' WHERE status IS NULL")
        conn.commit()
        print(f"Updated {cursor.rowcount} devices to have status = 'offline'")
    
    # Check the current status values
    cursor.execute("SELECT id, device_name, status FROM user_devices")
    devices = cursor.fetchall()
    
    print("\nCurrent device status:")
    for device in devices:
        print(f"  Device {device[1]} (ID: {device[0]}): status = '{device[2]}'")
    
    # Check if there are still pending messages
    cursor.execute("""
        SELECT COUNT(*) as count, bm.device_id, ud.device_name, ud.status
        FROM broadcast_messages bm
        LEFT JOIN user_devices ud ON bm.device_id = ud.id
        WHERE bm.status = 'pending'
        GROUP BY bm.device_id, ud.device_name, ud.status
    """)
    pending = cursor.fetchall()
    
    print("\nPending messages by device:")
    for row in pending:
        print(f"  Device {row[2]} (ID: {row[1]}): {row[0]} pending messages, status: {row[3]}")
    
    print("\nIssue should be fixed now! The broadcast processor should work properly.")
    
except mysql.connector.Error as err:
    print(f"Error: {err}")
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
