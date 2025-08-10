import psycopg2

DB_URI = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== Checking Messages Without Campaign or Sequence ID ===\n")
    
    # Check messages without campaign_id or sequence_id
    cur.execute("""
        SELECT COUNT(*) 
        FROM broadcast_messages 
        WHERE campaign_id IS NULL 
        AND sequence_id IS NULL
        AND status = 'pending'
    """)
    
    count = cur.fetchone()[0]
    print(f"Messages without campaign_id or sequence_id: {count}")
    
    if count > 0:
        print("\nSample messages with missing IDs:")
        cur.execute("""
            SELECT id, recipient_phone, message_type, 
                   campaign_id, sequence_id, sequence_stepid,
                   created_at
            FROM broadcast_messages 
            WHERE campaign_id IS NULL 
            AND sequence_id IS NULL
            AND status = 'pending'
            LIMIT 10
        """)
        
        for row in cur.fetchall():
            print(f"\nMessage ID: {row[0]}")
            print(f"  Phone: {row[1]}")
            print(f"  Type: {row[2]}")
            print(f"  Campaign ID: {row[3]}")
            print(f"  Sequence ID: {row[4]}")
            print(f"  Sequence Step ID: {row[5]}")
            print(f"  Created: {row[6]}")
    
    # Check sequence messages specifically
    print("\n=== Checking Sequence Messages ===")
    cur.execute("""
        SELECT COUNT(*),
               COUNT(CASE WHEN sequence_stepid IS NOT NULL THEN 1 END) as with_stepid
        FROM broadcast_messages 
        WHERE sequence_id IS NOT NULL
        AND status = 'pending'
    """)
    
    result = cur.fetchone()
    print(f"Total sequence messages: {result[0]}")
    print(f"With sequence_stepid: {result[1]}")
    
    # Check campaign messages
    print("\n=== Checking Campaign Messages ===")
    cur.execute("""
        SELECT COUNT(*)
        FROM broadcast_messages 
        WHERE campaign_id IS NOT NULL
        AND status = 'pending'
    """)
    
    count = cur.fetchone()[0]
    print(f"Total campaign messages: {count}")
    
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
