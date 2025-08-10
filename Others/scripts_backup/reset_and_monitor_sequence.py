import psycopg2
from datetime import datetime

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

try:
    # Connect to database
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("Connected to PostgreSQL database...")
    print("=" * 60)
    
    # Reset the sequence for testing
    print("\n1. Resetting WARM Sequence...")
    
    # Delete existing enrollments
    cur.execute("""
        DELETE FROM sequence_contacts
        WHERE sequence_id = (SELECT id FROM sequences WHERE name = 'WARM Sequence')
    """)
    print(f"Deleted {cur.rowcount} enrollment records")
    
    # Delete failed messages
    cur.execute("""
        DELETE FROM broadcast_messages
        WHERE sequence_id = (SELECT id FROM sequences WHERE name = 'WARM Sequence')
        AND status IN ('failed', 'skipped')
    """)
    print(f"Deleted {cur.rowcount} failed/skipped messages")
    
    # Commit
    conn.commit()
    
    print("\n2. Sequence has been reset!")
    print("\nWhat will happen next:")
    print("- Sequence processor will find 15 WARMEXAMA leads")
    print("- Create 4 records per lead (60 total)")
    print("- Only Step 1 will be 'active' for each lead")
    print("- Steps 2,3,4 will be 'pending' with proper delays")
    print("- Only active steps with next_trigger_time <= NOW will process")
    
    print("\n3. Run this query to monitor:")
    print("""
    SELECT 
        sc.contact_phone,
        ss.day_number,
        sc.status,
        sc.current_step,
        sc.next_trigger_time,
        sc.completed_at
    FROM sequence_contacts sc
    JOIN sequence_steps ss ON ss.id = sc.sequence_stepid
    WHERE sc.sequence_id = (SELECT id FROM sequences WHERE name = 'WARM Sequence')
    ORDER BY sc.contact_phone, ss.day_number
    LIMIT 20;
    """)
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
    if conn:
        conn.rollback()
        conn.close()
