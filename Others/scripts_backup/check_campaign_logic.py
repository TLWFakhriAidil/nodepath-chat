import psycopg2

conn_string = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(conn_string)
    cur = conn.cursor()
    
    print("=== CHECKING WHY CAMPAIGNS DON'T CREATE MESSAGES ===")
    
    # Get all devices for the user
    cur.execute("""
        SELECT id, phone, status, platform
        FROM user_devices 
        WHERE user_id = 'de078f16-3266-4ab3-8153-a248b015228f'
        AND (status = 'online' OR platform IS NOT NULL)
        ORDER BY platform
    """)
    
    devices = cur.fetchall()
    print(f"Found {len(devices)} connected devices")
    
    # Check leads for each device
    print("\n=== LEADS PER DEVICE ===")
    for device in devices:
        device_id = device[0]
        platform = device[3] if device[3] else "WhatsApp"
        
        # Count leads for this device
        cur.execute("""
            SELECT COUNT(*) 
            FROM leads 
            WHERE device_id = %s
        """, (device_id,))
        total_leads = cur.fetchone()[0]
        
        # Count leads matching campaign criteria
        cur.execute("""
            SELECT COUNT(*) 
            FROM leads 
            WHERE device_id = %s 
            AND niche = 'GRR' 
            AND target_status = 'prospect'
        """, (device_id,))
        matching_leads = cur.fetchone()[0]
        
        print(f"\nDevice: {device_id[:8]}... ({platform})")
        print(f"  Total leads: {total_leads}")
        print(f"  Matching leads (GRR/prospect): {matching_leads}")
        
        if matching_leads > 0:
            print(f"  ✓ This device SHOULD create {matching_leads} messages")
        else:
            print(f"  ✗ No matching leads - no messages will be created")
    
    print("\n=== THE PROBLEM ===")
    print("Platform devices (Wablas, Whacenter) have 0 leads!")
    print("Only your WhatsApp device has leads.")
    print("\nThe campaign processes ALL devices but only creates messages for devices with leads.")
    
    # Check the actual query used by GetLeadsByDeviceNicheAndStatus
    print("\n=== VERIFYING LEAD QUERY ===")
    cur.execute("""
        SELECT device_id, COUNT(*) as lead_count
        FROM leads 
        WHERE niche = 'GRR' 
        AND target_status = 'prospect'
        GROUP BY device_id
    """)
    
    print("Devices with matching leads:")
    for row in cur.fetchall():
        print(f"- Device {row[0]}: {row[1]} leads")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
