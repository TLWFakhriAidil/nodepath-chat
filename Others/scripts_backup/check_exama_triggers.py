import psycopg2

DB_URI = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== Checking EXAMA Triggers ===\n")
    
    # Check leads with EXAMA triggers
    print("Leads with EXAMA triggers:")
    cur.execute("""
        SELECT phone, trigger, device_id, user_id
        FROM leads 
        WHERE trigger LIKE '%EXAMA%'
        ORDER BY trigger
    """)
    
    leads = cur.fetchall()
    if leads:
        for lead in leads:
            print(f"  {lead[0]}: trigger='{lead[1]}'")
            print(f"    device_id: {lead[2]}")
            print(f"    user_id: {lead[3]}")
    else:
        print("  No leads found with EXAMA triggers!")
    
    print(f"\nTotal leads with EXAMA triggers: {len(leads)}")
    
    # Check what triggers exist
    print("\n\nAll unique triggers containing 'COLD', 'WARM', or 'HOT':")
    cur.execute("""
        SELECT DISTINCT trigger, COUNT(*) 
        FROM leads 
        WHERE trigger LIKE '%COLD%' OR trigger LIKE '%WARM%' OR trigger LIKE '%HOT%'
        GROUP BY trigger
        ORDER BY trigger
    """)
    
    triggers = cur.fetchall()
    for trig in triggers:
        print(f"  '{trig[0]}': {trig[1]} leads")
    
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
