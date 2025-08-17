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
    
    print("Fixing NULL device_status issue...")
    print("=" * 60)
    
    # First, check how many NULL values we have
    cursor.execute("SELECT COUNT(*) FROM broadcast_messages WHERE device_status IS NULL")
    null_count = cursor.fetchone()[0]
    print(f"Found {null_count} messages with NULL device_status")
    
    if null_count > 0:
        # Update NULL values to 'unknown'
        cursor.execute("UPDATE broadcast_messages SET device_status = 'unknown' WHERE device_status IS NULL")
        conn.commit()
        print(f"Updated {cursor.rowcount} rows to have device_status = 'unknown'")
    
    # Check the column definition
    cursor.execute("DESCRIBE broadcast_messages")
    columns = cursor.fetchall()
    for col in columns:
        if col[0] == 'device_status':
            print(f"\ndevice_status column: Type={col[1]}, Null={col[2]}, Default={col[4]}")
    
    # Check if there are pending messages
    cursor.execute("""
        SELECT COUNT(*) as pending_count, device_id 
        FROM broadcast_messages 
        WHERE status = 'pending' 
        GROUP BY device_id
    """)
    pending = cursor.fetchall()
    
    print(f"\nPending messages by device:")
    for row in pending:
        print(f"  Device {row[1]}: {row[0]} pending messages")
    
    print("\nIssue should be fixed now!")
    
except mysql.connector.Error as err:
    print(f"Error: {err}")
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
