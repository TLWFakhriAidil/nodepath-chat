import psycopg2

DB_URI = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== Updating Test Leads to Match New Triggers ===\n")
    
    # Update COLD EXSTART to COLDEXSTART
    cur.execute("""
        UPDATE leads
        SET trigger = 'COLDEXSTART'
        WHERE trigger = 'COLD EXSTART'
        RETURNING phone
    """)
    cold_phones = cur.fetchall()
    print(f"Updated {len(cold_phones)} leads from 'COLD EXSTART' to 'COLDEXSTART'")
    
    # Update HOT EXSTART to HOTEXSTART  
    cur.execute("""
        UPDATE leads
        SET trigger = 'HOTEXSTART'
        WHERE trigger = 'HOT EXSTART'
        RETURNING phone
    """)
    hot_phones = cur.fetchall()
    print(f"Updated {len(hot_phones)} leads from 'HOT EXSTART' to 'HOTEXSTART'")
    
    # For NEWNP EXSTART, first get some IDs then update
    cur.execute("""
        SELECT id FROM leads
        WHERE trigger = 'NEWNP EXSTART'
        AND device_id IS NOT NULL 
        AND user_id IS NOT NULL
        LIMIT 20
    """)
    
    lead_ids = [row[0] for row in cur.fetchall()]
    
    if lead_ids:
        # Convert list to tuple for SQL IN clause
        ids_tuple = tuple(lead_ids)
        cur.execute("""
            UPDATE leads
            SET trigger = 'COLDEXSTART'
            WHERE id IN %s
            RETURNING phone
        """, (ids_tuple,))
        test_phones = cur.fetchall()
        print(f"Updated {len(test_phones)} test leads from 'NEWNP EXSTART' to 'COLDEXSTART'")
    
    conn.commit()
    
    # Final check - how many leads match now
    print("\n=== Final Lead Counts ===")
    triggers = ['COLDEXSTART', 'WARMEXSTART', 'HOTEXSTART']
    
    for trigger in triggers:
        cur.execute("SELECT COUNT(*) FROM leads WHERE trigger = %s", (trigger,))
        count = cur.fetchone()[0]
        print(f"  Leads with '{trigger}': {count}")
    
    # Show summary of all sequences
    print("\n=== Sequences Summary ===")
    cur.execute("""
        SELECT s.name, s.is_active, ss.trigger as entry_trigger, 
               COUNT(DISTINCT l.phone) as matching_leads
        FROM sequences s
        JOIN sequence_steps ss ON s.id = ss.sequence_id AND ss.is_entry_point = true
        LEFT JOIN leads l ON l.trigger = ss.trigger
        GROUP BY s.name, s.is_active, ss.trigger
        ORDER BY s.name
    """)
    
    sequences = cur.fetchall()
    for seq in sequences:
        print(f"\n{seq[0]}:")
        print(f"  Active: {seq[1]}")
        print(f"  Entry trigger: {seq[2]}")
        print(f"  Matching leads: {seq[3]}")
    
    conn.close()
    print("\nDone! The system should now be able to enroll leads into sequences.")
    
except Exception as e:
    print(f"Error: {e}")
