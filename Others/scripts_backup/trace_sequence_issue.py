import psycopg2

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

try:
    # Connect to database
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("Connected to PostgreSQL database...")
    print("=" * 60)
    
    # Let's trace through what should happen
    print("\n1. Current state of sequence_contacts for one phone:")
    cur.execute("""
        SELECT 
            sc.id,
            sc.sequence_stepid,
            ss.day_number,
            sc.current_trigger,
            ss.next_trigger,
            sc.status,
            sc.next_trigger_time,
            sc.completed_at,
            sc.current_step
        FROM sequence_contacts sc
        JOIN sequence_steps ss ON ss.id = sc.sequence_stepid
        WHERE sc.contact_phone = '60199204337'
        ORDER BY ss.day_number
    """)
    
    results = cur.fetchall()
    print("Phone: 60199204337")
    print("\nStep-by-step records:")
    for row in results:
        print(f"\nDay {row[2]}:")
        print(f"  Record ID: {row[0][:8]}...")
        print(f"  Current Trigger: {row[3]}")
        print(f"  Next Trigger: {row[4]}")
        print(f"  Status: {row[5]}")
        print(f"  Next Trigger Time: {row[6]}")
        print(f"  Completed At: {row[7]}")
        print(f"  Current Step Value: {row[8]}")
    
    # The problem analysis
    print("\n\n2. THE PROBLEM:")
    print("All records show 'completed' status but:")
    print("- They were all completed at the same time (within seconds)")
    print("- The delays were not respected")
    print("- The 'current_step' field shows 4 for all (should be 1,2,3,4)")
    
    print("\n3. WHAT WENT WRONG:")
    print("After processing Step 1:")
    print("- It should mark Step 1 as 'sent' or 'completed'")
    print("- It should find Step 2 (WHERE current_trigger = 'WARMEXAMA_day2')")
    print("- It should update Step 2 from 'pending' to 'active'")
    print("- But this didn't happen correctly")
    
    # Check if the broadcast worker might have processed everything
    print("\n4. Checking broadcast worker behavior:")
    cur.execute("""
        SELECT 
            COUNT(*) as total,
            COUNT(DISTINCT device_id) as devices_used,
            MIN(created_at) as first_created,
            MAX(created_at) as last_created,
            STRING_AGG(DISTINCT status, ', ') as statuses
        FROM broadcast_messages
        WHERE sequence_id = (SELECT id FROM sequences WHERE name = 'WARM Sequence')
    """)
    
    broadcast_summary = cur.fetchone()
    print(f"Broadcast messages summary:")
    print(f"  Total messages: {broadcast_summary[0]}")
    print(f"  Devices used: {broadcast_summary[1]}")
    print(f"  First created: {broadcast_summary[2]}")
    print(f"  Last created: {broadcast_summary[3]}")
    print(f"  Statuses: {broadcast_summary[4]}")
    
    # The real issue
    print("\n5. THE REAL ISSUE:")
    print("The sequence processor likely:")
    print("1. Found all 'pending' records (because of a bug in the query)")
    print("2. Processed them all at once")
    print("3. Marked them all as 'completed'")
    print("4. Ignored the delay timings")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
    if conn:
        conn.close()
