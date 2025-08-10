import psycopg2

DB_URI = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== Fixing Messages: Adding sequence_id from sequence_stepid ===\n")
    
    # First, check how many messages need fixing
    cur.execute("""
        SELECT COUNT(*) 
        FROM broadcast_messages 
        WHERE sequence_stepid IS NOT NULL 
        AND sequence_id IS NULL
        AND status = 'pending'
    """)
    
    count = cur.fetchone()[0]
    print(f"Messages with sequence_stepid but no sequence_id: {count}")
    
    if count > 0:
        # Update sequence_id based on sequence_stepid
        print("\nUpdating sequence_id from sequence_steps table...")
        cur.execute("""
            UPDATE broadcast_messages bm
            SET sequence_id = ss.sequence_id
            FROM sequence_steps ss
            WHERE bm.sequence_stepid = ss.id
            AND bm.sequence_id IS NULL
            AND bm.sequence_stepid IS NOT NULL
        """)
        
        updated = cur.rowcount
        conn.commit()
        
        print(f"Updated {updated} messages with sequence_id!")
        
        # Verify the fix
        print("\n=== Verification ===")
        cur.execute("""
            SELECT 
                COUNT(*) as total,
                COUNT(sequence_id) as with_sequence_id,
                COUNT(sequence_stepid) as with_stepid,
                COUNT(CASE WHEN sequence_id IS NOT NULL AND sequence_stepid IS NOT NULL THEN 1 END) as with_both
            FROM broadcast_messages 
            WHERE status = 'pending'
        """)
        
        result = cur.fetchone()
        print(f"Pending messages:")
        print(f"  Total: {result[0]}")
        print(f"  With sequence_id: {result[1]}")
        print(f"  With sequence_stepid: {result[2]}")
        print(f"  With both IDs: {result[3]}")
        
        # Check if any are still missing sequence_id
        cur.execute("""
            SELECT COUNT(*) 
            FROM broadcast_messages 
            WHERE sequence_stepid IS NOT NULL 
            AND sequence_id IS NULL
        """)
        
        still_missing = cur.fetchone()[0]
        if still_missing > 0:
            print(f"\nWARNING: {still_missing} messages still missing sequence_id!")
            
            # Show why they're missing
            cur.execute("""
                SELECT DISTINCT bm.sequence_stepid
                FROM broadcast_messages bm
                LEFT JOIN sequence_steps ss ON bm.sequence_stepid = ss.id
                WHERE bm.sequence_stepid IS NOT NULL 
                AND bm.sequence_id IS NULL
                AND ss.id IS NULL
                LIMIT 5
            """)
            
            orphaned = cur.fetchall()
            if orphaned:
                print("These sequence_stepid values don't exist in sequence_steps table:")
                for step in orphaned:
                    print(f"  - {step[0]}")
    
    conn.close()
    print("\nDone! Messages should now have both sequence_id and sequence_stepid.")
    
except Exception as e:
    print(f"Error: {e}")
