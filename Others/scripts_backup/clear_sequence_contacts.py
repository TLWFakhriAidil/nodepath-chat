# Script to remove all sequence_contacts records from PostgreSQL

import psycopg2

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

try:
    # Connect to database
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    # First, let's see how many records we have
    cur.execute("SELECT COUNT(*) FROM sequence_contacts")
    count = cur.fetchone()[0]
    print(f"Found {count} sequence_contacts records")
    
    if count > 0:
        # Delete all records
        cur.execute("DELETE FROM sequence_contacts")
        deleted_count = cur.rowcount
        conn.commit()
        print(f"Successfully deleted {deleted_count} sequence_contacts records")
    else:
        print("No records to delete")
    
    # Verify deletion
    cur.execute("SELECT COUNT(*) FROM sequence_contacts")
    new_count = cur.fetchone()[0]
    print(f"Records remaining: {new_count}")
    
    cur.close()
    conn.close()
    print("Done!")
    
except Exception as e:
    print(f"Error: {e}")
