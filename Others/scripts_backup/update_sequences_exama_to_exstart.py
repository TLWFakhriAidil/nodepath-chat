import psycopg2

DB_URI = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== Updating Sequences Table from EXAMA to EXSTART ===\n")
    
    # First, show current triggers in sequences table
    print("Current sequences with EXAMA triggers:")
    cur.execute("""
        SELECT id, name, trigger, start_trigger, end_trigger
        FROM sequences
        WHERE trigger LIKE '%EXAMA%' 
           OR start_trigger LIKE '%EXAMA%' 
           OR end_trigger LIKE '%EXAMA%'
    """)
    
    sequences = cur.fetchall()
    for seq in sequences:
        print(f"  {seq[1]}:")
        print(f"    trigger: {seq[2]}")
        print(f"    start_trigger: {seq[3]}")
        print(f"    end_trigger: {seq[4]}")
    
    print(f"\nTotal sequences to update: {len(sequences)}")
    
    # Update trigger column
    print("\nUpdating 'trigger' column...")
    cur.execute("""
        UPDATE sequences
        SET trigger = REPLACE(trigger, 'EXAMA', 'EXSTART')
        WHERE trigger LIKE '%EXAMA%'
        RETURNING name, trigger
    """)
    
    updated = cur.fetchall()
    for row in updated:
        print(f"  {row[0]}: trigger updated to '{row[1]}'")
    
    # Update start_trigger column
    print("\nUpdating 'start_trigger' column...")
    cur.execute("""
        UPDATE sequences
        SET start_trigger = REPLACE(start_trigger, 'EXAMA', 'EXSTART')
        WHERE start_trigger LIKE '%EXAMA%'
        RETURNING name, start_trigger
    """)
    
    updated = cur.fetchall()
    for row in updated:
        print(f"  {row[0]}: start_trigger updated to '{row[1]}'")
    
    # Update end_trigger column
    print("\nUpdating 'end_trigger' column...")
    cur.execute("""
        UPDATE sequences
        SET end_trigger = REPLACE(end_trigger, 'EXAMA', 'EXSTART')
        WHERE end_trigger LIKE '%EXAMA%'
        RETURNING name, end_trigger
    """)
    
    updated = cur.fetchall()
    for row in updated:
        print(f"  {row[0]}: end_trigger updated to '{row[1]}'")
    
    # Commit changes
    conn.commit()
    
    # Verify the updates
    print("\n=== Verifying Updates ===")
    cur.execute("""
        SELECT name, trigger, start_trigger, end_trigger, is_active
        FROM sequences
        ORDER BY name
    """)
    
    print("All sequences after update:")
    sequences = cur.fetchall()
    for seq in sequences:
        print(f"\n{seq[0]} (Active: {seq[4]}):")
        print(f"  trigger: {seq[1]}")
        print(f"  start_trigger: {seq[2]}")
        print(f"  end_trigger: {seq[3]}")
    
    # Now update some leads to match the new triggers
    print("\n=== Updating Test Leads ===")
    
    # Update COLD EXSTART to COLDEXSTART
    cur.execute("""
        UPDATE leads
        SET trigger = 'COLDEXSTART'
        WHERE trigger = 'COLD EXSTART'
        RETURNING phone
    """)
    cold_count = len(cur.fetchall())
    
    # Update HOT EXSTART to HOTEXSTART  
    cur.execute("""
        UPDATE leads
        SET trigger = 'HOTEXSTART'
        WHERE trigger = 'HOT EXSTART'
        RETURNING phone
    """)
    hot_count = len(cur.fetchall())
    
    # Update some NEWNP EXSTART to COLDEXSTART for testing
    cur.execute("""
        UPDATE leads
        SET trigger = 'COLDEXSTART'
        WHERE trigger = 'NEWNP EXSTART'
        AND device_id IS NOT NULL 
        AND user_id IS NOT NULL
        LIMIT 20
        RETURNING phone
    """)
    test_count = len(cur.fetchall())
    
    conn.commit()
    
    print(f"\nUpdated {cold_count} leads from 'COLD EXSTART' to 'COLDEXSTART'")
    print(f"Updated {hot_count} leads from 'HOT EXSTART' to 'HOTEXSTART'")
    print(f"Updated {test_count} test leads from 'NEWNP EXSTART' to 'COLDEXSTART'")
    
    # Final check - how many leads match now
    print("\n=== Final Lead Counts ===")
    triggers = ['COLDEXSTART', 'WARMEXSTART', 'HOTEXSTART']
    
    for trigger in triggers:
        cur.execute("SELECT COUNT(*) FROM leads WHERE trigger = %s", (trigger,))
        count = cur.fetchone()[0]
        print(f"  Leads with '{trigger}': {count}")
    
    conn.close()
    print("\nSuccessfully updated all sequences and some test leads!")
    
except Exception as e:
    print(f"Error: {e}")
