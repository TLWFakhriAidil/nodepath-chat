import psycopg2
from datetime import datetime, timedelta

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

try:
    # Connect to database
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("Connected to PostgreSQL database...")
    print("=" * 60)
    
    # Option 1: Reset the sequence to try again
    print("\n1. Resetting WARM Sequence for testing...")
    
    # Delete all existing sequence_contacts for WARM sequence
    cur.execute("""
        DELETE FROM sequence_contacts
        WHERE sequence_id = (SELECT id FROM sequences WHERE name = 'WARM Sequence')
        RETURNING contact_phone
    """)
    
    deleted_phones = cur.fetchall()
    print(f"Deleted {len(deleted_phones)} enrollment records")
    
    # Delete failed broadcast messages
    cur.execute("""
        DELETE FROM broadcast_messages
        WHERE sequence_id = (SELECT id FROM sequences WHERE name = 'WARM Sequence')
        AND status = 'failed'
    """)
    print(f"Deleted {cur.rowcount} failed messages")
    
    # Commit the cleanup
    conn.commit()
    
    print("\n2. Ready for re-enrollment!")
    print("The sequence processor should now:")
    print("- Find the 15 WARMEXAMA leads again")
    print("- Create 4 records per lead (60 total)")
    print("- Only process Step 1 immediately")
    print("- Wait 12 hours for Step 2, etc.")
    
    print("\n3. To monitor proper processing:")
    print("Watch for:")
    print("- Only 'active' status records being processed")
    print("- Proper delays between steps")
    print("- One step at a time per contact")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
    if conn:
        conn.rollback()
        conn.close()
