import psycopg2

DB_URI = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== Setting up test leads with EXAMA triggers ===\n")
    
    # Update some leads to have COLDEXAMA trigger for testing
    test_phones = ['601126152825', '60122631139', '60142818860']
    
    updated_count = 0
    for phone in test_phones:
        cur.execute("""
            UPDATE leads 
            SET trigger = 'COLDEXAMA'
            WHERE phone = %s 
            AND device_id IS NOT NULL 
            AND user_id IS NOT NULL
            RETURNING phone, trigger, device_id
        """, (phone,))
        
        result = cur.fetchone()
        if result:
            print(f"Updated {result[0]} to trigger 'COLDEXAMA'")
            updated_count += 1
    
    conn.commit()
    print(f"\nTotal leads updated: {updated_count}")
    
    # Verify the updates
    print("\n=== Verifying EXAMA triggers ===")
    cur.execute("""
        SELECT phone, trigger, device_id IS NOT NULL as has_device, user_id IS NOT NULL as has_user
        FROM leads 
        WHERE trigger LIKE '%EXAMA%'
        ORDER BY trigger
    """)
    
    leads = cur.fetchall()
    for lead in leads:
        print(f"  {lead[0]}: trigger='{lead[1]}', has_device={lead[2]}, has_user={lead[3]}")
    
    print(f"\nTotal leads with EXAMA triggers: {len(leads)}")
    
    conn.close()
    print("\nTest leads are ready for Direct Broadcast enrollment!")
    
except Exception as e:
    print(f"Error: {e}")
