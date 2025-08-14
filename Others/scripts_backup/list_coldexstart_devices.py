import psycopg2

DB_URI = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== Devices with COLDEXSTART Leads ===\n")
    
    # Get devices with COLDEXSTART leads
    cur.execute("""
        SELECT 
            l.device_id,
            ud.device_name,
            ud.platform,
            ud.status,
            COUNT(DISTINCT l.phone) as lead_count,
            STRING_AGG(DISTINCT l.phone, ', ' ORDER BY l.phone) as sample_phones
        FROM leads l
        JOIN user_devices ud ON l.device_id = ud.id
        WHERE l.trigger = 'COLDEXSTART'
        GROUP BY l.device_id, ud.device_name, ud.platform, ud.status
        ORDER BY lead_count DESC
    """)
    
    devices = cur.fetchall()
    
    if devices:
        print(f"Found {len(devices)} device(s) with COLDEXSTART leads:\n")
        
        for device in devices:
            print(f"Device ID: {device[0]}")
            print(f"  Name: {device[1]}")
            print(f"  Platform: {device[2]}")
            print(f"  Status: {device[3]}")
            print(f"  Lead Count: {device[4]}")
            print(f"  Sample Phones: {device[5][:100]}{'...' if len(device[5]) > 100 else ''}")
            print()
    else:
        print("No devices found with COLDEXSTART leads")
    
    # Also show detailed lead info
    print("\n=== All COLDEXSTART Leads by Device ===")
    cur.execute("""
        SELECT 
            l.phone,
            l.name,
            l.device_id,
            l.user_id,
            ud.device_name,
            ud.status as device_status
        FROM leads l
        LEFT JOIN user_devices ud ON l.device_id = ud.id
        WHERE l.trigger = 'COLDEXSTART'
        ORDER BY l.device_id, l.phone
    """)
    
    leads = cur.fetchall()
    current_device = None
    device_count = 0
    
    for lead in leads:
        if current_device != lead[2]:
            device_count += 1
            current_device = lead[2]
            print(f"\nDevice #{device_count}: {lead[4]} (ID: {lead[2]})")
            print(f"Status: {lead[5]}")
            print("Leads:")
        
        print(f"  - {lead[0]} ({lead[1] or 'No name'})")
    
    print(f"\n=== Summary ===")
    print(f"Total COLDEXSTART leads: {len(leads)}")
    print(f"Total unique devices: {len(devices)}")
    
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
