import psycopg2

DB_URI = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    # Test inserting with empty UUID string
    print("=== Testing UUID insert scenarios ===")
    
    # This should fail
    try:
        cur.execute("SELECT ''::uuid")
        print("Empty string to UUID: SUCCESS (shouldn't happen!)")
    except Exception as e:
        print(f"Empty string to UUID: FAILED - {e}")
    
    # This should work
    try:
        cur.execute("SELECT NULL::uuid")
        print("NULL to UUID: SUCCESS")
    except Exception as e:
        print(f"NULL to UUID: FAILED - {e}")
    
    # Check if any user_devices have empty user_id
    print("\n=== Checking user_devices for empty user_id ===")
    cur.execute("""
        SELECT id, user_id, platform 
        FROM user_devices 
        WHERE user_id IS NULL OR user_id::text = ''
        LIMIT 5
    """)
    
    if cur.rowcount > 0:
        print("Found devices with empty user_id:")
        for row in cur.fetchall():
            print(f"  device_id: {row[0]}, user_id: '{row[1]}', platform: {row[2]}")
    else:
        print("No devices with empty user_id")
    
    # Check specific device
    print("\n=== Checking specific device ===")
    cur.execute("""
        SELECT id, user_id, status 
        FROM user_devices 
        WHERE id = '315e4f8e-6868-4808-a3df-f75e9fce331f'
    """)
    
    result = cur.fetchone()
    if result:
        print(f"Device: {result[0]}")
        print(f"user_id: '{result[1]}'")
        print(f"status: {result[2]}")
    
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
