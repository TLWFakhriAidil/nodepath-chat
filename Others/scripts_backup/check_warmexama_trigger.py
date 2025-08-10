import psycopg2

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

try:
    # Connect to database
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("Connected to PostgreSQL database...")
    print("=" * 60)
    
    # Check leads with trigger WARMEXAMA
    print("\nChecking leads with trigger 'WARMEXAMA'...")
    
    # Count exact matches
    cur.execute("""
        SELECT COUNT(*) 
        FROM leads 
        WHERE trigger = 'WARMEXAMA'
    """)
    exact_count = cur.fetchone()[0]
    print(f"\n1. Exact match (trigger = 'WARMEXAMA'): {exact_count} leads")
    
    # Count partial matches (in case trigger contains WARMEXAMA along with other values)
    cur.execute("""
        SELECT COUNT(*) 
        FROM leads 
        WHERE trigger LIKE '%WARMEXAMA%'
    """)
    partial_count = cur.fetchone()[0]
    print(f"\n2. Partial match (trigger contains 'WARMEXAMA'): {partial_count} leads")
    
    # Show some examples
    print("\n3. Sample leads with WARMEXAMA trigger:")
    cur.execute("""
        SELECT name, phone, device_id, niche, trigger, created_at
        FROM leads 
        WHERE trigger LIKE '%WARMEXAMA%'
        ORDER BY created_at DESC
        LIMIT 10
    """)
    
    samples = cur.fetchall()
    for name, phone, device_id, niche, trigger, created_at in samples:
        print(f"   Name: {name}, Phone: {phone}, Niche: {niche}")
        print(f"   Trigger: {trigger}")
        print(f"   Device: {device_id[:8]}..., Created: {created_at}")
        print()
    
    # Check distribution by device
    print("\n4. Distribution by device:")
    cur.execute("""
        SELECT d.device_name, COUNT(l.id) as lead_count
        FROM leads l
        JOIN user_devices d ON l.device_id = d.id
        WHERE l.trigger LIKE '%WARMEXAMA%'
        GROUP BY d.device_name
        ORDER BY lead_count DESC
        LIMIT 10
    """)
    
    device_dist = cur.fetchall()
    for device_name, lead_count in device_dist:
        print(f"   {device_name}: {lead_count} leads")
    
    # Check if combined with other triggers
    print("\n5. Checking if WARMEXAMA is combined with other triggers:")
    cur.execute("""
        SELECT DISTINCT trigger
        FROM leads 
        WHERE trigger LIKE '%WARMEXAMA%' AND trigger != 'WARMEXAMA'
        LIMIT 10
    """)
    
    combined = cur.fetchall()
    if combined:
        print("   Found leads with multiple triggers:")
        for (trigger,) in combined:
            print(f"   - {trigger}")
    else:
        print("   No combined triggers found - all are exact 'WARMEXAMA'")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
    if conn:
        conn.close()
