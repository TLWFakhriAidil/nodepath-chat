import psycopg2

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

try:
    # Connect to database
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("Connected to PostgreSQL database...")
    print("=" * 60)
    
    # First, find both devices
    print("\n1. Finding devices...")
    
    # Find chedin device
    cur.execute("""
        SELECT id, user_id, device_name, jid, platform
        FROM user_devices 
        WHERE device_name = 'chedin'
    """)
    chedin = cur.fetchone()
    
    if not chedin:
        print("ERROR: Device 'chedin' not found!")
        exit()
    
    chedin_id, user_id, _, chedin_jid, platform = chedin
    print(f"Found 'chedin' device: {chedin_id}")
    
    # Find SCHQ-S09 device
    cur.execute("""
        SELECT id, device_name, jid
        FROM user_devices 
        WHERE device_name = 'SCHQ-S09' AND user_id = %s
    """, (user_id,))
    schq = cur.fetchone()
    
    if not schq:
        print("ERROR: Device 'SCHQ-S09' not found for the same user!")
        exit()
    
    schq_id, _, schq_jid = schq
    print(f"Found 'SCHQ-S09' device: {schq_id}")
    
    # Count leads to transfer
    cur.execute("""
        SELECT COUNT(*) 
        FROM leads 
        WHERE device_id = %s
    """, (chedin_id,))
    lead_count = cur.fetchone()[0]
    print(f"\n2. Found {lead_count} leads to transfer from 'chedin' to 'SCHQ-S09'")
    
    # Update all leads from chedin to SCHQ-S09
    print("\n3. Transferring leads...")
    cur.execute("""
        UPDATE leads 
        SET device_id = %s, updated_at = NOW()
        WHERE device_id = %s
        RETURNING id
    """, (schq_id, chedin_id))
    
    transferred = cur.fetchall()
    print(f"Transferred {len(transferred)} leads successfully")
    
    # Update SCHQ-S09 with chedin's JID if it's newer/different
    print("\n4. Updating SCHQ-S09 device...")
    if chedin_jid and chedin_jid != schq_jid:
        print(f"Updating SCHQ-S09 JID from '{schq_jid}' to '{chedin_jid}'")
        cur.execute("""
            UPDATE user_devices 
            SET jid = %s, updated_at = NOW()
            WHERE id = %s
        """, (chedin_jid, schq_id))
    
    # Delete chedin device
    print("\n5. Deleting 'chedin' device...")
    cur.execute("""
        DELETE FROM user_devices 
        WHERE id = %s
        RETURNING device_name
    """, (chedin_id,))
    
    deleted = cur.fetchone()
    if deleted:
        print(f"Deleted device: {deleted[0]}")
    
    # Commit all changes
    conn.commit()
    
    # Verify the transfer
    print("\n6. Verification:")
    
    # Check SCHQ-S09 leads
    cur.execute("""
        SELECT COUNT(*) 
        FROM leads l
        JOIN user_devices d ON l.device_id = d.id
        WHERE d.device_name = 'SCHQ-S09'
    """)
    final_count = cur.fetchone()[0]
    print(f"SCHQ-S09 now has {final_count} total leads")
    
    # Check no more chedin
    cur.execute("""
        SELECT COUNT(*) 
        FROM user_devices 
        WHERE device_name = 'chedin'
    """)
    chedin_remaining = cur.fetchone()[0]
    print(f"Remaining 'chedin' devices: {chedin_remaining}")
    
    print("\n" + "=" * 60)
    print("MERGE COMPLETED SUCCESSFULLY!")
    print(f"- Transferred {len(transferred)} leads from 'chedin' to 'SCHQ-S09'")
    print(f"- Deleted 'chedin' device")
    print(f"- SCHQ-S09 now has all the leads")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
    if conn:
        conn.rollback()
        conn.close()
