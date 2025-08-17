import psycopg2

conn_string = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(conn_string)
    cur = conn.cursor()
    
    print("=== CHECKING WHY CAMPAIGN 59 IS NOT PROCESSING ===")
    
    # Check if campaign has any broadcast messages (even deleted ones)
    cur.execute("""
        SELECT COUNT(*) 
        FROM broadcast_messages 
        WHERE campaign_id = 59
    """)
    msg_count = cur.fetchone()[0]
    print(f"Broadcast messages for campaign 59: {msg_count}")
    
    # Check the exact query used by GetPendingCampaigns
    cur.execute("""
        SELECT 
            id, title, status, campaign_date, time_schedule,
            (campaign_date || ' ' || time_schedule)::TIMESTAMP AT TIME ZONE 'Asia/Kuala_Lumpur' as scheduled_time,
            CURRENT_TIMESTAMP as current_time,
            CASE 
                WHEN id IN (SELECT DISTINCT campaign_id FROM broadcast_messages WHERE campaign_id IS NOT NULL) 
                THEN 'Has broadcast messages - EXCLUDED'
                ELSE 'No broadcast messages'
            END as exclusion_reason
        FROM campaigns
        WHERE id = 59
    """)
    
    result = cur.fetchone()
    if result:
        print(f"\nCampaign ID: {result[0]}")
        print(f"Title: {result[1]}")
        print(f"Status: {result[2]}")
        print(f"Campaign Date: {result[3]}")
        print(f"Time Schedule: {result[4]}")
        print(f"Scheduled Time (KL): {result[5]}")
        print(f"Current Time: {result[6]}")
        print(f"Exclusion Reason: {result[7]}")
    
    # The fix: Update campaign status to trigger it
    print("\n=== SOLUTION ===")
    print("The campaign processor is excluding this campaign because it might have old messages.")
    print("To fix this, we need to:")
    print("1. Delete any old broadcast messages for this campaign")
    print("2. Or update the campaign status to 'scheduled' and back to 'pending'")
    
    # Check if we should clean up
    cur.execute("SELECT COUNT(*) FROM broadcast_messages WHERE campaign_id = 59")
    if cur.fetchone()[0] > 0:
        print("\nFound old broadcast messages. Cleaning up...")
        cur.execute("DELETE FROM broadcast_messages WHERE campaign_id = 59")
        conn.commit()
        print("Deleted old messages. Campaign should process now.")
    else:
        print("\nNo old messages found. Checking campaign trigger query...")
        
        # Run the exact query from GetPendingCampaigns
        cur.execute("""
            SELECT id, title 
            FROM campaigns
            WHERE status = 'pending'
            AND id NOT IN (
                SELECT DISTINCT campaign_id 
                FROM broadcast_messages 
                WHERE campaign_id IS NOT NULL
            )
            AND (
                time_schedule IS NULL 
                OR time_schedule = ''
                OR (campaign_date || ' ' || time_schedule)::TIMESTAMP AT TIME ZONE 'Asia/Kuala_Lumpur' <= CURRENT_TIMESTAMP
            )
        """)
        
        print("\nCampaigns that SHOULD process:")
        for row in cur.fetchall():
            print(f"- ID: {row[0]}, Title: {row[1]}")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
