import psycopg2

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

# Connect to database
conn = psycopg2.connect(DB_URI)
cur = conn.cursor()

try:
    # Delete all records from broadcast_messages
    cur.execute("DELETE FROM broadcast_messages")
    broadcast_deleted = cur.rowcount
    print(f"Deleted {broadcast_deleted} records from broadcast_messages")
    
    # Delete all records from sequence_contacts
    cur.execute("DELETE FROM sequence_contacts")
    sequence_deleted = cur.rowcount
    print(f"Deleted {sequence_deleted} records from sequence_contacts")
    
    # Commit the changes
    conn.commit()
    print("\nDeletion completed successfully!")
    
    # Verify deletion
    cur.execute("SELECT COUNT(*) FROM broadcast_messages")
    broadcast_count = cur.fetchone()[0]
    
    cur.execute("SELECT COUNT(*) FROM sequence_contacts")
    sequence_count = cur.fetchone()[0]
    
    print(f"\nVerification:")
    print(f"broadcast_messages: {broadcast_count} records (should be 0)")
    print(f"sequence_contacts: {sequence_count} records (should be 0)")
    
except Exception as e:
    conn.rollback()
    print(f"Error: {e}")
finally:
    cur.close()
    conn.close()
