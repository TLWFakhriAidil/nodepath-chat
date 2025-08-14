import psycopg2

DB_URI = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== Checking Broadcast Messages Status Values ===\n")
    
    # Get all unique status values
    cur.execute("""
        SELECT status, COUNT(*) as count
        FROM broadcast_messages
        GROUP BY status
        ORDER BY COUNT(*) DESC
    """)
    
    statuses = cur.fetchall()
    print("All status values in broadcast_messages:")
    print("-" * 40)
    for status, count in statuses:
        print(f"{status:<20} : {count:,} messages")
    
    # Check if 'success' status exists
    cur.execute("""
        SELECT COUNT(*) 
        FROM broadcast_messages 
        WHERE status = 'success'
    """)
    
    success_count = cur.fetchone()[0]
    
    if success_count > 0:
        print(f"\nFound {success_count} messages with status='success'")
        
        # Show some samples
        print("\nSample success messages:")
        cur.execute("""
            SELECT id, recipient_phone, sent_at, campaign_id, sequence_stepid
            FROM broadcast_messages 
            WHERE status = 'success'
            LIMIT 5
        """)
        
        for row in cur.fetchall():
            print(f"  ID: {row[0]}, Phone: {row[1]}, Sent: {row[2]}")
    else:
        print("\nNo messages with status='success' found")
    
    # Check for messages that were actually sent
    print("\n=== Checking Sent Messages ===")
    cur.execute("""
        SELECT status, COUNT(*) 
        FROM broadcast_messages 
        WHERE sent_at IS NOT NULL
        GROUP BY status
    """)
    
    sent_statuses = cur.fetchall()
    if sent_statuses:
        print("Messages with sent_at timestamp:")
        for status, count in sent_statuses:
            print(f"  Status '{status}': {count} messages")
    else:
        print("No messages have sent_at timestamp")
    
    # Check expected status values
    print("\n=== Expected Status Values ===")
    expected = ['pending', 'sent', 'failed', 'success', 'delivered']
    for status in expected:
        cur.execute("SELECT COUNT(*) FROM broadcast_messages WHERE status = %s", (status,))
        count = cur.fetchone()[0]
        if count > 0:
            print(f"  '{status}': {count} messages")
    
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
