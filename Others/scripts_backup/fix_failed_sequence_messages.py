import psycopg2

DB_URI = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== Fixing Failed Messages with sequence_stepid ===\n")
    
    # First, count how many messages need fixing
    cur.execute("""
        SELECT COUNT(*) 
        FROM broadcast_messages 
        WHERE status = 'failed'
        AND error_message = 'message has no campaign or sequence ID'
        AND sequence_stepid IS NOT NULL
    """)
    
    count = cur.fetchone()[0]
    print(f"Found {count} failed messages to fix")
    
    if count > 0:
        # Show some samples before fixing
        print("\nSample messages to be fixed:")
        cur.execute("""
            SELECT id, recipient_phone, sequence_stepid, error_message
            FROM broadcast_messages 
            WHERE status = 'failed'
            AND error_message = 'message has no campaign or sequence ID'
            AND sequence_stepid IS NOT NULL
            LIMIT 5
        """)
        
        for row in cur.fetchall():
            print(f"  ID: {row[0]}, Phone: {row[1]}, Step: {row[2]}")
        
        # Update all failed messages back to pending
        print(f"\nUpdating {count} messages back to pending status...")
        cur.execute("""
            UPDATE broadcast_messages 
            SET status = 'pending',
                error_message = NULL
            WHERE status = 'failed'
            AND error_message = 'message has no campaign or sequence ID'
            AND sequence_stepid IS NOT NULL
        """)
        
        updated = cur.rowcount
        conn.commit()
        
        print(f"✅ Successfully updated {updated} messages to pending status!")
        
        # Verify the update
        cur.execute("""
            SELECT COUNT(*) 
            FROM broadcast_messages 
            WHERE status = 'pending'
            AND sequence_stepid IS NOT NULL
            AND campaign_id IS NULL
            AND sequence_id IS NULL
        """)
        
        pending_count = cur.fetchone()[0]
        print(f"\nTotal pending messages with only sequence_stepid: {pending_count}")
    else:
        print("No failed messages found with this specific error.")
    
    # Also check for any other failed messages with sequence_stepid
    print("\n=== Checking Other Failed Messages ===")
    cur.execute("""
        SELECT error_message, COUNT(*) 
        FROM broadcast_messages 
        WHERE status = 'failed'
        AND sequence_stepid IS NOT NULL
        GROUP BY error_message
        ORDER BY COUNT(*) DESC
        LIMIT 10
    """)
    
    other_errors = cur.fetchall()
    if other_errors:
        print("Other errors for sequence messages:")
        for error, count in other_errors:
            print(f"  '{error}': {count} messages")
    else:
        print("No other failed sequence messages found.")
    
    conn.close()
    print("\n✅ Database fix completed!")
    
except Exception as e:
    print(f"Error: {e}")
