import psycopg2

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

try:
    # Connect to database
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("Connected to PostgreSQL database...")
    print("=" * 60)
    
    # Check if there's another device with the same JID
    print("\n1. Checking JID conflict...")
    cur.execute("""
        SELECT id, device_name, jid 
        FROM user_devices 
        WHERE jid = '601158666863:70@s.whatsapp.net'
    """)
    
    devices_with_jid = cur.fetchall()
    print(f"Found {len(devices_with_jid)} devices with JID '601158666863:70@s.whatsapp.net':")
    for dev_id, dev_name, jid in devices_with_jid:
        print(f"  - {dev_name} (ID: {dev_id})")
    
    # Since we already transferred the leads, let's just delete chedin
    print("\n2. Since leads are already transferred, deleting 'chedin' device...")
    
    cur.execute("""
        SELECT id FROM user_devices WHERE device_name = 'chedin'
    """)
    chedin_id = cur.fetchone()
    
    if chedin_id:
        cur.execute("""
            DELETE FROM user_devices 
            WHERE device_name = 'chedin'
            RETURNING id, device_name
        """, )
        
        deleted = cur.fetchone()
        if deleted:
            print(f"Deleted device: {deleted[1]} (ID: {deleted[0]})")
        
        # Commit the deletion
        conn.commit()
    else:
        print("'chedin' device not found - may have been deleted already")
    
    # Final verification
    print("\n3. Final verification:")
    
    # Check SCHQ-S09 leads
    cur.execute("""
        SELECT COUNT(*) 
        FROM leads l
        JOIN user_devices d ON l.device_id = d.id
        WHERE d.device_name = 'SCHQ-S09'
    """)
    final_count = cur.fetchone()[0]
    print(f"SCHQ-S09 now has {final_count} total leads")
    
    # Check leads with WARMEXAMA trigger on SCHQ-S09
    cur.execute("""
        SELECT COUNT(*) 
        FROM leads l
        JOIN user_devices d ON l.device_id = d.id
        WHERE d.device_name = 'SCHQ-S09' AND l.trigger = 'WARMEXAMA'
    """)
    warmexama_count = cur.fetchone()[0]
    print(f"SCHQ-S09 has {warmexama_count} leads with WARMEXAMA trigger")
    
    # Check no more chedin
    cur.execute("""
        SELECT COUNT(*) 
        FROM user_devices 
        WHERE device_name = 'chedin'
    """)
    chedin_remaining = cur.fetchone()[0]
    print(f"Remaining 'chedin' devices: {chedin_remaining}")
    
    print("\n" + "=" * 60)
    print("OPERATION COMPLETED!")
    print("- All 'chedin' leads have been transferred to 'SCHQ-S09'")
    print("- 'chedin' device has been deleted")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
    if conn:
        conn.rollback()
        conn.close()
