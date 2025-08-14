import psycopg2

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

try:
    # Connect to database
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("Connected to PostgreSQL database...")
    print("=" * 60)
    
    # Check active sequences
    print("\n1. Checking active sequences...")
    cur.execute("""
        SELECT id, name, trigger, is_active
        FROM sequences
        WHERE is_active = true
    """)
    
    sequences = cur.fetchall()
    print(f"Found {len(sequences)} active sequences:")
    for seq_id, name, trigger, is_active in sequences:
        print(f"  - {name} (Trigger: {trigger}, ID: {seq_id})")
    
    # Check sequence steps
    print("\n2. Checking sequence steps...")
    for seq_id, name, _, _ in sequences:
        cur.execute("""
            SELECT id, day_number, trigger, is_entry_point
            FROM sequence_steps
            WHERE sequence_id = %s
            ORDER BY day_number
            LIMIT 5
        """, (seq_id,))
        
        steps = cur.fetchall()
        if steps:
            print(f"\n  Steps for '{name}':")
            for step_id, day, trigger, entry in steps:
                print(f"    Day {day}: {trigger} (Entry: {entry}, ID: {step_id})")
    
    # Check leads with WARMEXAMA trigger
    print("\n3. Checking leads ready for sequence enrollment...")
    cur.execute("""
        SELECT COUNT(*)
        FROM leads
        WHERE trigger = 'WARMEXAMA'
    """)
    warmexama_count = cur.fetchone()[0]
    print(f"Leads with WARMEXAMA trigger: {warmexama_count}")
    
    # Check if there's a sequence matching WARMEXAMA
    cur.execute("""
        SELECT s.name, ss.id, ss.trigger
        FROM sequences s
        JOIN sequence_steps ss ON ss.sequence_id = s.id
        WHERE ss.trigger = 'WARMEXAMA' AND ss.is_entry_point = true
    """)
    
    matching_seq = cur.fetchall()
    if matching_seq:
        print("\nSequences with WARMEXAMA entry point:")
        for name, step_id, trigger in matching_seq:
            print(f"  - {name} (Step ID: {step_id})")
    else:
        print("\nNo sequences found with WARMEXAMA as entry point")
    
    # Check current sequence_contacts
    print("\n4. Checking current sequence enrollments...")
    cur.execute("""
        SELECT COUNT(*), status
        FROM sequence_contacts
        GROUP BY status
    """)
    
    enrollments = cur.fetchall()
    if enrollments:
        print("Current enrollments by status:")
        for count, status in enrollments:
            print(f"  - {status}: {count}")
    else:
        print("No enrollments found")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
    if conn:
        conn.close()
