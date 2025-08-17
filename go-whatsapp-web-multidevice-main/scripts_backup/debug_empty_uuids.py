import psycopg2

DB_URI = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    # Check the specific failing leads
    failing_phones = ['601119241561', '601118696369']
    
    print("=== Checking Failing Leads ===")
    for phone in failing_phones:
        cur.execute("""
            SELECT l.phone, l.device_id, l.user_id, l.trigger,
                   ud.id as device_table_id, ud.user_id as device_user_id
            FROM leads l
            LEFT JOIN user_devices ud ON l.device_id = ud.id
            WHERE l.phone = %s
        """, (phone,))
        
        results = cur.fetchall()
        if results:
            print(f"\nPhone: {phone}")
            for row in results:
                print(f"  Lead device_id: '{row[1]}'")
                print(f"  Lead user_id: '{row[2]}'")
                print(f"  Lead trigger: {row[3]}")
                print(f"  Device exists: {'Yes' if row[4] else 'No'}")
                if row[4]:
                    print(f"  Device user_id: '{row[5]}'")
                
                # Check if values are empty strings
                if row[1] == '':
                    print("  ⚠️  Lead has EMPTY device_id!")
                if row[2] == '':
                    print("  ⚠️  Lead has EMPTY user_id!")
                if row[4] and row[5] == '':
                    print("  ⚠️  Device has EMPTY user_id!")
    
    # Check for any leads with empty UUIDs
    print("\n=== Checking All Leads with Empty UUIDs ===")
    cur.execute("""
        SELECT COUNT(*) FROM leads 
        WHERE device_id = '' OR user_id = ''
    """)
    count = cur.fetchone()[0]
    print(f"Leads with empty device_id or user_id: {count}")
    
    if count > 0:
        cur.execute("""
            SELECT phone, device_id, user_id, trigger
            FROM leads 
            WHERE device_id = '' OR user_id = ''
            LIMIT 10
        """)
        print("\nSample leads with empty UUIDs:")
        for row in cur.fetchall():
            print(f"  Phone: {row[0]}, device_id: '{row[1]}', user_id: '{row[2]}', trigger: {row[3]}")
    
    # Check user_devices for empty user_id
    print("\n=== Checking User Devices ===")
    cur.execute("""
        SELECT COUNT(*) FROM user_devices 
        WHERE user_id = '' OR user_id IS NULL
    """)
    count = cur.fetchone()[0]
    print(f"Devices with empty or NULL user_id: {count}")
    
    if count > 0:
        cur.execute("""
            SELECT id, device_name, user_id, platform
            FROM user_devices 
            WHERE user_id = '' OR user_id IS NULL
            LIMIT 5
        """)
        print("\nDevices with empty user_id:")
        for row in cur.fetchall():
            print(f"  ID: {row[0]}, Name: {row[1]}, user_id: '{row[2]}', Platform: {row[3]}")
    
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
