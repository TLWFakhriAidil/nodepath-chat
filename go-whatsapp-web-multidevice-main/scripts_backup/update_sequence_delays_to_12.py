import psycopg2

DB_URI = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== Updating All Sequence Steps trigger_delay_hours to 12 ===\n")
    
    # First, show current trigger_delay_hours
    print("Current trigger_delay_hours values:")
    cur.execute("""
        SELECT s.name, ss.trigger, ss.trigger_delay_hours, ss.day_number
        FROM sequence_steps ss
        JOIN sequences s ON ss.sequence_id = s.id
        ORDER BY s.name, ss.day_number
    """)
    
    steps = cur.fetchall()
    current_sequence = None
    
    for step in steps:
        if current_sequence != step[0]:
            current_sequence = step[0]
            print(f"\n{current_sequence}:")
        print(f"  Day {step[3]}: {step[1]} - delay: {step[2]} hours")
    
    print(f"\nTotal steps to update: {len(steps)}")
    
    # Update all trigger_delay_hours to 12
    print("\nUpdating all steps to 12 hour delay...")
    cur.execute("""
        UPDATE sequence_steps
        SET trigger_delay_hours = 12
        RETURNING id, trigger
    """)
    
    updated = cur.fetchall()
    print(f"Updated {len(updated)} steps")
    
    # Commit changes
    conn.commit()
    
    # Verify the updates
    print("\n=== Verifying Updates ===")
    cur.execute("""
        SELECT s.name, ss.trigger, ss.trigger_delay_hours, ss.day_number
        FROM sequence_steps ss
        JOIN sequences s ON ss.sequence_id = s.id
        ORDER BY s.name, ss.day_number
    """)
    
    steps = cur.fetchall()
    current_sequence = None
    
    for step in steps:
        if current_sequence != step[0]:
            current_sequence = step[0]
            print(f"\n{current_sequence}:")
        print(f"  Day {step[3]}: {step[1]} - delay: {step[2]} hours")
    
    # Calculate total time for each sequence
    print("\n=== Sequence Timing Summary ===")
    cur.execute("""
        SELECT s.name, COUNT(*) as steps, SUM(ss.trigger_delay_hours) as total_hours
        FROM sequences s
        JOIN sequence_steps ss ON s.id = ss.sequence_id
        GROUP BY s.name
        ORDER BY s.name
    """)
    
    sequences = cur.fetchall()
    for seq in sequences:
        days = seq[2] / 24.0
        print(f"{seq[0]}: {seq[1]} steps, {seq[2]} hours ({days:.1f} days)")
    
    conn.close()
    print("\nDone! All sequence steps now have 12 hour delays between messages.")
    
except Exception as e:
    print(f"Error: {e}")
