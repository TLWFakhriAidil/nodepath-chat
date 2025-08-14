import psycopg2

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

try:
    # Connect to database
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("Connected to PostgreSQL database...")
    print("=" * 60)
    
    # Check what happened to the leads
    print("\n1. Checking lead status after transfer...")
    
    # Total leads with WARMEXAMA
    cur.execute("""
        SELECT COUNT(*) 
        FROM leads 
        WHERE trigger = 'WARMEXAMA'
    """)
    total_warmexama = cur.fetchone()[0]
    print(f"Total leads with WARMEXAMA trigger: {total_warmexama}")
    
    # Check if there are orphaned leads (device_id pointing to deleted device)
    cur.execute("""
        SELECT COUNT(*) 
        FROM leads l
        LEFT JOIN user_devices d ON l.device_id = d.id
        WHERE d.id IS NULL
    """)
    orphaned = cur.fetchone()[0]
    print(f"\nOrphaned leads (no device): {orphaned}")
    
    # Find the SCHQ-S09 device ID
    cur.execute("""
        SELECT id, device_name 
        FROM user_devices 
        WHERE device_name = 'SCHQ-S09'
    """)
    schq_device = cur.fetchone()
    if schq_device:
        schq_id = schq_device[0]
        print(f"\nSCHQ-S09 device ID: {schq_id}")
        
        # Count leads for SCHQ-S09
        cur.execute("""
            SELECT COUNT(*) 
            FROM leads 
            WHERE device_id = %s
        """, (schq_id,))
        schq_leads = cur.fetchone()[0]
        print(f"Leads assigned to SCHQ-S09: {schq_leads}")
    
    # Check all devices with WARMEXAMA leads
    print("\n2. Distribution of WARMEXAMA leads by device:")
    cur.execute("""
        SELECT 
            COALESCE(d.device_name, 'NO DEVICE') as device_name,
            d.id as device_id,
            COUNT(l.id) as lead_count
        FROM leads l
        LEFT JOIN user_devices d ON l.device_id = d.id
        WHERE l.trigger = 'WARMEXAMA'
        GROUP BY d.device_name, d.id
        ORDER BY lead_count DESC
    """)
    
    distribution = cur.fetchall()
    for device_name, device_id, lead_count in distribution:
        print(f"  {device_name}: {lead_count} leads (ID: {device_id})")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
    if conn:
        conn.close()
