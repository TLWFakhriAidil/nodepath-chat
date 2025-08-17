import psycopg2
import sys

# Database connection
conn_string = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"
device_id = "d409cadc-75e2-4004-a789-c2bad0b31393"

try:
    conn = psycopg2.connect(conn_string)
    cur = conn.cursor()
    
    print("=== DEVICE STATUS CHECK ===")
    cur.execute("SELECT id, status, platform FROM user_devices WHERE id = %s", (device_id,))
    result = cur.fetchone()
    if result:
        print(f"Device ID: {result[0]}")
        print(f"Status: '{result[1]}'")
        print(f"Platform: '{result[2]}'")
    else:
        print("Device not found!")
    
    print("\n=== ALL DEVICE STATUSES ===")
    cur.execute("SELECT DISTINCT status, COUNT(*) FROM user_devices GROUP BY status ORDER BY status")
    for row in cur.fetchall():
        print(f"Status '{row[0]}': {row[1]} devices")
    
    print("\n=== CAMPAIGN 59 CHECK ===")
    cur.execute("SELECT id, title, status FROM campaigns WHERE id = 59")
    result = cur.fetchone()
    if result:
        print(f"Campaign: {result[1]}")
        print(f"Status: {result[2]}")
    
    print("\n=== BROADCAST MESSAGES STATUS ===")
    cur.execute("""
        SELECT status, COUNT(*) 
        FROM broadcast_messages 
        WHERE device_id = %s 
        GROUP BY status
    """, (device_id,))
    for row in cur.fetchall():
        print(f"{row[0]}: {row[1]} messages")
    
    print("\n=== CAMPAIGN 59 MESSAGES ===")
    cur.execute("""
        SELECT status, COUNT(*) 
        FROM broadcast_messages 
        WHERE campaign_id = 59 
        GROUP BY status
    """)
    for row in cur.fetchall():
        print(f"{row[0]}: {row[1]} messages")
    
    print("\n=== RECENT MESSAGES (Last 5) ===")
    cur.execute("""
        SELECT id, device_id, campaign_id, status, error_message, created_at
        FROM broadcast_messages 
        WHERE campaign_id = 59 OR device_id = %s
        ORDER BY created_at DESC 
        LIMIT 5
    """, (device_id,))
    
    for row in cur.fetchall():
        print(f"\nID: {row[0][:8]}...")
        print(f"  Device: {row[1]}")
        print(f"  Campaign: {row[2]}")
        print(f"  Status: {row[3]}")
        print(f"  Error: {row[4]}")
        print(f"  Created: {row[5]}")
    
    # Check why messages might be skipped
    print("\n=== CHECKING WHY MESSAGES AREN'T PROCESSING ===")
    cur.execute("""
        SELECT b.status as msg_status, d.status as device_status, b.scheduled_at, b.created_at
        FROM broadcast_messages b
        JOIN user_devices d ON b.device_id = d.id
        WHERE b.campaign_id = 59 AND b.status = 'pending'
        LIMIT 5
    """)
    
    for row in cur.fetchall():
        print(f"Message status: {row[0]}, Device status: {row[1]}")
        print(f"Scheduled: {row[2]}, Created: {row[3]}")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
