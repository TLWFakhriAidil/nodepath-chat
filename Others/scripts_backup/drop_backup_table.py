import psycopg2

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

# Connect to database
conn = psycopg2.connect(DB_URI)
cur = conn.cursor()

try:
    # Drop the backup table
    cur.execute("DROP TABLE IF EXISTS sequence_contacts_backup_fix")
    
    # Commit the changes
    conn.commit()
    print("Table 'sequence_contacts_backup_fix' has been dropped successfully!")
    
    # Check if there are any other backup tables
    cur.execute("""
        SELECT table_name 
        FROM information_schema.tables 
        WHERE table_schema = 'public' 
        AND table_name LIKE '%backup%'
        ORDER BY table_name
    """)
    
    backup_tables = cur.fetchall()
    
    if backup_tables:
        print("\nOther backup tables found:")
        for table in backup_tables:
            print(f"  - {table[0]}")
    else:
        print("\nNo other backup tables found.")
    
except Exception as e:
    conn.rollback()
    print(f"Error: {e}")
finally:
    cur.close()
    conn.close()
