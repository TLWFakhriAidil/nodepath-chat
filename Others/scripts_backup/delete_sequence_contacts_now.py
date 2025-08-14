import psycopg2

# Connect to database
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"
conn = psycopg2.connect(DB_URI)
cur = conn.cursor()

print("Connected to PostgreSQL!")

# Check current count
cur.execute("SELECT COUNT(*) FROM sequence_contacts")
total = cur.fetchone()[0]
print(f"\nCurrent total sequence contacts: {total}")

# Check breakdown by status
print("\nBreakdown by status:")
cur.execute("SELECT status, COUNT(*) FROM sequence_contacts GROUP BY status")
for row in cur.fetchall():
    print(f"  {row[0]}: {row[1]}")

if total > 0:
    # DELETE ALL RECORDS
    print("\nDeleting all sequence contacts...")
    cur.execute("DELETE FROM sequence_contacts")
    deleted = cur.rowcount
    
    # Commit the deletion
    conn.commit()
    print(f"\nSuccessfully deleted {deleted} records")
    
    # Verify deletion
    cur.execute("SELECT COUNT(*) FROM sequence_contacts")
    remaining = cur.fetchone()[0]
    print(f"Remaining contacts: {remaining}")
else:
    print("\nNo records to delete.")

# Close connection
cur.close()
conn.close()
print("\nDone!")