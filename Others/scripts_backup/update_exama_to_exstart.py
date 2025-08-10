import psycopg2

DB_URI = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== Updating Sequence Steps from EXAMA to EXSTART ===\n")
    
    # First, show current triggers with EXAMA
    print("Current triggers containing EXAMA:")
    cur.execute("""
        SELECT id, sequence_id, trigger, next_trigger, is_entry_point
        FROM sequence_steps
        WHERE trigger LIKE '%EXAMA%' OR next_trigger LIKE '%EXAMA%'
        ORDER BY sequence_id, day_number
    """)
    
    steps = cur.fetchall()
    for step in steps:
        print(f"  Step {step[0]}: trigger='{step[2]}', next_trigger='{step[3]}', entry_point={step[4]}")
    
    print(f"\nTotal steps to update: {len(steps)}")
    
    # Update triggers
    print("\nUpdating triggers...")
    cur.execute("""
        UPDATE sequence_steps
        SET trigger = REPLACE(trigger, 'EXAMA', 'EXSTART')
        WHERE trigger LIKE '%EXAMA%'
        RETURNING id, trigger
    """)
    
    updated_triggers = cur.fetchall()
    for row in updated_triggers:
        print(f"  Updated trigger to: {row[1]}")
    
    # Update next_triggers
    print("\nUpdating next_triggers...")
    cur.execute("""
        UPDATE sequence_steps
        SET next_trigger = REPLACE(next_trigger, 'EXAMA', 'EXSTART')
        WHERE next_trigger LIKE '%EXAMA%'
        RETURNING id, next_trigger
    """)
    
    updated_next_triggers = cur.fetchall()
    for row in updated_next_triggers:
        print(f"  Updated next_trigger to: {row[1]}")
    
    # Commit changes
    conn.commit()
    
    # Verify the updates
    print("\n=== Verifying Updates ===")
    cur.execute("""
        SELECT DISTINCT trigger
        FROM sequence_steps
        WHERE is_entry_point = true
        ORDER BY trigger
    """)
    
    print("Entry point triggers after update:")
    for row in cur.fetchall():
        print(f"  {row[0]}")
    
    # Show how many leads match new triggers
    print("\n=== Checking Lead Matches ===")
    triggers = ['COLDEXSTART', 'WARMEXSTART', 'HOTEXSTART']
    
    for trigger in triggers:
        cur.execute("""
            SELECT COUNT(*) 
            FROM leads 
            WHERE trigger = %s
        """, (trigger,))
        count = cur.fetchone()[0]
        print(f"  Leads with '{trigger}': {count}")
    
    # Check partial matches
    print("\nChecking partial matches:")
    cur.execute("""
        SELECT COUNT(*) 
        FROM leads 
        WHERE trigger LIKE '%EXSTART%'
    """)
    count = cur.fetchone()[0]
    print(f"  Total leads containing 'EXSTART': {count}")
    
    conn.close()
    print("\n✅ Successfully updated all EXAMA to EXSTART in sequence steps!")
    
except Exception as e:
    print(f"Error: {e}")
