import psycopg2

DB_URI = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== ALL TRIGGERS IN LEADS TABLE (Grouped by Count) ===\n")
    
    # Get all triggers grouped by count
    cur.execute("""
        SELECT 
            trigger,
            COUNT(*) as lead_count,
            COUNT(DISTINCT device_id) as unique_devices,
            COUNT(DISTINCT user_id) as unique_users
        FROM leads 
        WHERE trigger IS NOT NULL AND trigger != ''
        GROUP BY trigger
        ORDER BY COUNT(*) DESC
    """)
    
    triggers = cur.fetchall()
    
    print(f"{'Trigger':<30} {'Count':<10} {'Devices':<10} {'Users':<10}")
    print("-" * 60)
    
    total_leads = 0
    for trigger in triggers:
        print(f"{trigger[0]:<30} {trigger[1]:<10} {trigger[2]:<10} {trigger[3]:<10}")
        total_leads += trigger[1]
    
    print("-" * 60)
    print(f"Total unique triggers: {len(triggers)}")
    print(f"Total leads with triggers: {total_leads}")
    
    # Also show triggers that might match sequences
    print("\n\n=== TRIGGERS CONTAINING SEQUENCE KEYWORDS ===")
    print("\nTriggers with 'EXAMA':")
    cur.execute("""
        SELECT trigger, COUNT(*) 
        FROM leads 
        WHERE trigger LIKE '%EXAMA%'
        GROUP BY trigger
        ORDER BY COUNT(*) DESC
    """)
    
    for row in cur.fetchall():
        print(f"  {row[0]}: {row[1]} leads")
    
    print("\nTriggers with 'COLD':")
    cur.execute("""
        SELECT trigger, COUNT(*) 
        FROM leads 
        WHERE trigger LIKE '%COLD%'
        GROUP BY trigger
        ORDER BY COUNT(*) DESC
    """)
    
    for row in cur.fetchall():
        print(f"  {row[0]}: {row[1]} leads")
    
    print("\nTriggers with 'WARM':")
    cur.execute("""
        SELECT trigger, COUNT(*) 
        FROM leads 
        WHERE trigger LIKE '%WARM%'
        GROUP BY trigger
        ORDER BY COUNT(*) DESC
    """)
    
    for row in cur.fetchall():
        print(f"  {row[0]}: {row[1]} leads")
    
    print("\nTriggers with 'HOT':")
    cur.execute("""
        SELECT trigger, COUNT(*) 
        FROM leads 
        WHERE trigger LIKE '%HOT%'
        GROUP BY trigger
        ORDER BY COUNT(*) DESC
    """)
    
    for row in cur.fetchall():
        print(f"  {row[0]}: {row[1]} leads")
    
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
