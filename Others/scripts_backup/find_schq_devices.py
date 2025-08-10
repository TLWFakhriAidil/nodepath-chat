import psycopg2

DB_URI = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== Searching for SCHQ devices ===\n")
    
    # Search for SCHQ devices
    cur.execute("""
        SELECT device_name, id, platform, status,
               (SELECT COUNT(*) FROM leads WHERE device_id = ud.id) as lead_count
        FROM user_devices ud
        WHERE device_name LIKE 'SCHQ%' OR device_name LIKE 'SCAS-S09'
        ORDER BY device_name
    """)
    
    devices = cur.fetchall()
    
    if devices:
        print(f"Found {len(devices)} SCHQ/similar devices:")
        for dev in devices:
            print(f"\n{dev[0]}:")
            print(f"  ID: {dev[1]}")
            print(f"  Platform: {dev[2]}")
            print(f"  Status: {dev[3]}")
            print(f"  Total leads: {dev[4]}")
    
    # Let's use SCAS-S09 which we saw earlier
    print("\n\n=== Using SCAS-S09 device ===")
    cur.execute("""
        SELECT id, device_name, platform, status, user_id
        FROM user_devices
        WHERE device_name = 'SCAS-S09'
    """)
    
    device = cur.fetchone()
    if device:
        device_id = device[0]
        print(f"Device: {device[1]}")
        print(f"  ID: {device_id}")
        print(f"  Status: {device[3]}")
        
        # Get leads
        cur.execute("""
            SELECT phone, name, trigger
            FROM leads
            WHERE device_id = %s
            ORDER BY phone
        """, (device_id,))
        
        leads = cur.fetchall()
        print(f"\nTotal leads: {len(leads)}")
        
        # Show current trigger distribution
        cur.execute("""
            SELECT trigger, COUNT(*) 
            FROM leads
            WHERE device_id = %s
            GROUP BY trigger
        """, (device_id,))
        
        print("\nCurrent triggers:")
        for row in cur.fetchall():
            print(f"  {row[0]}: {row[1]} leads")
    
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
