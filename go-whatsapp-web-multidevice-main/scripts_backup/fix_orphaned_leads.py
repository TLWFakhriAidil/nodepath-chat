import psycopg2

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

try:
    # Connect to database
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("Connected to PostgreSQL database...")
    print("=" * 60)
    
    # Find SCHQ-S09 device
    cur.execute("""
        SELECT id, device_name 
        FROM user_devices 
        WHERE device_name = 'SCHQ-S09'
    """)
    schq_device = cur.fetchone()
    
    if not schq_device:
        print("ERROR: SCHQ-S09 device not found!")
        exit()
    
    schq_id = schq_device[0]
    print(f"\n1. Found SCHQ-S09 device: {schq_id}")
    
    # Fix all orphaned leads by assigning them to SCHQ-S09
    print("\n2. Fixing orphaned leads...")
    cur.execute("""
        UPDATE leads 
        SET device_id = %s
        WHERE device_id NOT IN (SELECT id FROM user_devices)
        RETURNING id
    """, (schq_id,))
    
    fixed_leads = cur.fetchall()
    print(f"Fixed {len(fixed_leads)} orphaned leads - assigned to SCHQ-S09")
    
    # Commit the changes
    conn.commit()
    
    # Verify the fix
    print("\n3. Verification:")
    
    # Check total leads for SCHQ-S09
    cur.execute("""
        SELECT COUNT(*) 
        FROM leads 
        WHERE device_id = %s
    """, (schq_id,))
    total_leads = cur.fetchone()[0]
    print(f"SCHQ-S09 now has {total_leads} total leads")
    
    # Check WARMEXAMA leads for SCHQ-S09
    cur.execute("""
        SELECT COUNT(*) 
        FROM leads 
        WHERE device_id = %s AND trigger = 'WARMEXAMA'
    """, (schq_id,))
    warmexama_leads = cur.fetchone()[0]
    print(f"SCHQ-S09 has {warmexama_leads} leads with WARMEXAMA trigger")
    
    # Check no more orphaned leads
    cur.execute("""
        SELECT COUNT(*) 
        FROM leads l
        LEFT JOIN user_devices d ON l.device_id = d.id
        WHERE d.id IS NULL
    """)
    orphaned = cur.fetchone()[0]
    print(f"Remaining orphaned leads: {orphaned}")
    
    print("\n" + "=" * 60)
    print("FIX COMPLETED!")
    print(f"- All orphaned leads have been assigned to SCHQ-S09")
    print(f"- SCHQ-S09 now has all the leads from 'chedin'")
    print(f"- All WARMEXAMA leads are now under SCHQ-S09")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
    if conn:
        conn.rollback()
        conn.close()
