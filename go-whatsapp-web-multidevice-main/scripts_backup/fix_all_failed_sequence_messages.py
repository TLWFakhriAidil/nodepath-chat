import psycopg2

DB_URI = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== Fixing All Failed Sequence Messages ===\n")
    
    # Fix both old and new error messages
    error_messages = [
        'message has no campaign or sequence ID',
        'message has no campaign ID or sequence step ID'
    ]
    
    total_fixed = 0
    
    for error_msg in error_messages:
        # Count messages with this error
        cur.execute("""
            SELECT COUNT(*) 
            FROM broadcast_messages 
            WHERE status = 'failed'
            AND error_message = %s
            AND sequence_stepid IS NOT NULL
        """, (error_msg,))
        
        count = cur.fetchone()[0]
        
        if count > 0:
            print(f"Found {count} failed messages with error: '{error_msg}'")
            
            # Update them back to pending
            cur.execute("""
                UPDATE broadcast_messages 
                SET status = 'pending',
                    error_message = NULL
                WHERE status = 'failed'
                AND error_message = %s
                AND sequence_stepid IS NOT NULL
            """, (error_msg,))
            
            updated = cur.rowcount
            total_fixed += updated
            print(f"  Updated {updated} messages to pending\n")
    
    # Also fix any messages that have sequence_stepid but no campaign_id/sequence_id
    print("=== Fixing Orphaned Messages ===")
    cur.execute("""
        UPDATE broadcast_messages 
        SET status = 'pending',
            error_message = NULL
        WHERE status IN ('failed', 'error')
        AND sequence_stepid IS NOT NULL
        AND campaign_id IS NULL
        AND sequence_id IS NULL
    """)
    
    orphan_fixed = cur.rowcount
    total_fixed += orphan_fixed
    
    if orphan_fixed > 0:
        print(f"Fixed {orphan_fixed} orphaned messages")
    
    # Commit all changes
    conn.commit()
    
    print(f"\n=== Summary ===")
    print(f"Total messages fixed: {total_fixed}")
    
    # Show current status
    cur.execute("""
        SELECT 
            COUNT(*) as total,
            COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending,
            COUNT(CASE WHEN status = 'sent' THEN 1 END) as sent,
            COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed
        FROM broadcast_messages 
        WHERE sequence_stepid IS NOT NULL
        AND campaign_id IS NULL
        AND sequence_id IS NULL
    """)
    
    stats = cur.fetchone()
    print(f"\nMessages with only sequence_stepid:")
    print(f"  Total: {stats[0]}")
    print(f"  Pending: {stats[1]}")
    print(f"  Sent: {stats[2]}")
    print(f"  Failed: {stats[3]}")
    
    conn.close()
    print("\nAll failed sequence messages have been reset to pending!")
    
except Exception as e:
    print(f"Error: {e}")
