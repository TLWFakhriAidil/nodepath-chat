import psycopg2

conn_string = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(conn_string)
    cur = conn.cursor()
    
    print("=== COMPARING SEQUENCES vs CAMPAIGNS ===")
    
    # Check sequence messages
    cur.execute("""
        SELECT COUNT(*), MIN(created_at), MAX(created_at)
        FROM broadcast_messages 
        WHERE sequence_id IS NOT NULL
        AND created_at > NOW() - INTERVAL '1 day'
    """)
    seq_count, seq_min, seq_max = cur.fetchone()
    print(f"\nSequence Messages (last 24h):")
    print(f"  Count: {seq_count}")
    print(f"  First: {seq_min}")
    print(f"  Last: {seq_max}")
    
    # Check campaign messages
    cur.execute("""
        SELECT COUNT(*), MIN(created_at), MAX(created_at)
        FROM broadcast_messages 
        WHERE campaign_id IS NOT NULL
        AND created_at > NOW() - INTERVAL '1 day'
    """)
    camp_count, camp_min, camp_max = cur.fetchone()
    print(f"\nCampaign Messages (last 24h):")
    print(f"  Count: {camp_count}")
    print(f"  First: {camp_min}")
    print(f"  Last: {camp_max}")
    
    # Check why campaigns might not be creating messages
    print("\n=== CHECKING CAMPAIGN EXECUTION ===")
    
    # Get recent campaigns
    cur.execute("""
        SELECT id, title, status, updated_at
        FROM campaigns 
        WHERE updated_at > NOW() - INTERVAL '1 hour'
        ORDER BY updated_at DESC
        LIMIT 5
    """)
    
    print("\nRecent campaigns:")
    for row in cur.fetchall():
        print(f"  ID: {row[0]}, Title: {row[1]}, Status: {row[2]}, Updated: {row[3]}")
    
    # Check the campaign log pattern
    print("\n=== THE KEY DIFFERENCE ===")
    print("Sequences: Create messages directly with ALL required fields")
    print("Campaigns: Might be failing silently during QueueMessage")
    
    # Check if there are any errors in broadcast_messages
    cur.execute("""
        SELECT device_id, status, error_message, COUNT(*)
        FROM broadcast_messages 
        WHERE campaign_id IN (59, 60)
        GROUP BY device_id, status, error_message
    """)
    
    print("\nCampaign 59/60 message status:")
    for row in cur.fetchall():
        print(f"  Device: {row[0][:8]}..., Status: {row[1]}, Error: {row[2]}, Count: {row[3]}")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
