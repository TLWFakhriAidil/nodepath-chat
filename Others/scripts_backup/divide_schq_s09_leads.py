import psycopg2

DB_URI = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== Checking Device SCHQ-S09 ===\n")
    
    # First check if device exists
    cur.execute("""
        SELECT id, device_name, platform, status, user_id
        FROM user_devices
        WHERE device_name = 'SCHQ-S09'
    """)
    
    device = cur.fetchone()
    if not device:
        print("Device SCHQ-S09 not found!")
        
        # Let's search for similar names
        cur.execute("""
            SELECT device_name, id, platform, status
            FROM user_devices
            WHERE device_name LIKE 'SCHQ%'
            ORDER BY device_name
        """)
        
        similar = cur.fetchall()
        if similar:
            print("\nSimilar devices found:")
            for dev in similar:
                print(f"  - {dev[0]} (ID: {dev[1]}, Platform: {dev[2]}, Status: {dev[3]})")
    else:
        device_id = device[0]
        print(f"Found device: {device[1]}")
        print(f"  ID: {device_id}")
        print(f"  Platform: {device[2]}")
        print(f"  Status: {device[3]}")
        print(f"  User ID: {device[4]}")
        
        # Get all leads for this device
        cur.execute("""
            SELECT phone, name, trigger, niche
            FROM leads
            WHERE device_id = %s
            ORDER BY phone
        """, (device_id,))
        
        leads = cur.fetchall()
        print(f"\nTotal leads for this device: {len(leads)}")
        
        if leads:
            # Calculate quarters
            total = len(leads)
            quarter = total // 4
            remainder = total % 4
            
            cold_count = quarter + (1 if remainder > 0 else 0)
            warm_count = quarter + (1 if remainder > 1 else 0)
            hot_count = quarter + (1 if remainder > 2 else 0)
            keep_count = total - cold_count - warm_count - hot_count
            
            print(f"\nDividing {total} leads:")
            print(f"  COLDEXSTART: {cold_count} leads")
            print(f"  WARMEXSTART: {warm_count} leads")
            print(f"  HOTEXSTART: {hot_count} leads")
            print(f"  Keep original: {keep_count} leads")
            
            # Update triggers
            print("\nUpdating triggers...")
            
            # Update first quarter to COLDEXSTART
            for i in range(cold_count):
                cur.execute("""
                    UPDATE leads
                    SET trigger = 'COLDEXSTART'
                    WHERE phone = %s AND device_id = %s
                """, (leads[i][0], device_id))
            
            # Update second quarter to WARMEXSTART
            for i in range(cold_count, cold_count + warm_count):
                cur.execute("""
                    UPDATE leads
                    SET trigger = 'WARMEXSTART'
                    WHERE phone = %s AND device_id = %s
                """, (leads[i][0], device_id))
            
            # Update third quarter to HOTEXSTART
            for i in range(cold_count + warm_count, cold_count + warm_count + hot_count):
                cur.execute("""
                    UPDATE leads
                    SET trigger = 'HOTEXSTART'
                    WHERE phone = %s AND device_id = %s
                """, (leads[i][0], device_id))
            
            conn.commit()
            
            # Verify the update
            cur.execute("""
                SELECT trigger, COUNT(*) 
                FROM leads
                WHERE device_id = %s
                GROUP BY trigger
                ORDER BY trigger
            """, (device_id,))
            
            print("\nUpdated trigger distribution:")
            for row in cur.fetchall():
                print(f"  {row[0]}: {row[1]} leads")
                
            # Show some sample leads
            print("\nSample leads after update:")
            cur.execute("""
                SELECT phone, name, trigger
                FROM leads
                WHERE device_id = %s 
                AND trigger IN ('COLDEXSTART', 'WARMEXSTART', 'HOTEXSTART')
                ORDER BY trigger, phone
                LIMIT 15
            """, (device_id,))
            
            for lead in cur.fetchall():
                print(f"  {lead[0]} ({lead[1] or 'No name'}) - {lead[2]}")
    
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
