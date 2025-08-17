import psycopg2

DB_URI = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== Checking Sequence Enrollment Issues ===\n")
    
    # 1. Check active sequences with entry points
    print("1. Active sequences with entry points:")
    cur.execute("""
        SELECT s.id, s.name, ss.trigger, count(ss.id) as step_count
        FROM sequences s
        JOIN sequence_steps ss ON s.id = ss.sequence_id
        WHERE s.is_active = true AND ss.is_entry_point = true
        GROUP BY s.id, s.name, ss.trigger
    """)
    
    sequences = cur.fetchall()
    for seq in sequences:
        print(f"   - {seq[1]}: trigger='{seq[2]}', steps={seq[3]}")
    
    # 2. Check leads with triggers
    print("\n2. Leads with triggers:")
    cur.execute("""
        SELECT trigger, COUNT(*) 
        FROM leads 
        WHERE trigger IS NOT NULL AND trigger != ''
        AND device_id IS NOT NULL AND user_id IS NOT NULL
        GROUP BY trigger
        ORDER BY COUNT(*) DESC
    """)
    
    triggers = cur.fetchall()
    for trig in triggers:
        print(f"   - {trig[0]}: {trig[1]} leads")
    
    # 3. Check if leads already have messages
    print("\n3. Checking enrollment blockers:")
    cur.execute("""
        SELECT l.phone, l.trigger, 
               (SELECT COUNT(*) FROM broadcast_messages bm 
                WHERE bm.recipient_phone = l.phone 
                AND bm.status IN ('pending', 'sent')) as msg_count
        FROM leads l
        WHERE l.trigger IS NOT NULL AND l.trigger != ''
        AND l.device_id IS NOT NULL AND l.user_id IS NOT NULL
        LIMIT 10
    """)
    
    leads = cur.fetchall()
    for lead in leads:
        print(f"   - {lead[0]}: trigger='{lead[1]}', existing_messages={lead[2]}")
    
    # 4. Check exact matching
    print("\n4. Testing trigger matching:")
    cur.execute("""
        SELECT DISTINCT l.phone, l.trigger, ss.trigger as step_trigger,
               position(ss.trigger in l.trigger) as match_pos
        FROM leads l
        CROSS JOIN sequences s
        INNER JOIN sequence_steps ss ON ss.sequence_id = s.id
        WHERE s.is_active = true 
        AND ss.is_entry_point = true
        AND l.trigger IS NOT NULL AND l.trigger != ''
        AND l.device_id IS NOT NULL AND l.user_id IS NOT NULL
        LIMIT 10
    """)
    
    matches = cur.fetchall()
    for match in matches:
        print(f"   - Lead {match[0]}: '{match[1]}' vs Step '{match[2]}' = position {match[3]}")
        if match[3] > 0:
            print(f"     ✅ MATCH!")
    
    # 5. Final eligibility check
    print("\n5. Final eligibility query:")
    cur.execute("""
        SELECT COUNT(DISTINCT l.phone)
        FROM leads l
        CROSS JOIN sequences s
        INNER JOIN sequence_steps ss ON ss.sequence_id = s.id
        WHERE s.is_active = true 
        AND ss.is_entry_point = true
        AND l.trigger IS NOT NULL AND l.trigger != ''
        AND l.device_id IS NOT NULL AND l.user_id IS NOT NULL
        AND position(ss.trigger in l.trigger) > 0
        AND NOT EXISTS (
            SELECT 1 FROM broadcast_messages bm
            WHERE bm.sequence_id = s.id 
            AND bm.recipient_phone = l.phone
            AND bm.status IN ('pending', 'sent')
        )
    """)
    
    eligible = cur.fetchone()[0]
    print(f"   Total eligible leads for enrollment: {eligible}")
    
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
