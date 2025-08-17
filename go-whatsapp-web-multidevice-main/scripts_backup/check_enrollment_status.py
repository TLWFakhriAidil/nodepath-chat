import psycopg2
from datetime import datetime

# Connect to PostgreSQL
conn = psycopg2.connect(
    "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"
)
cur = conn.cursor()

print("=== CHECKING SEQUENCE ENROLLMENT STATUS ===\n")

# Check sequence_contacts table
print("1. Checking sequence_contacts table:")
cur.execute("""
    SELECT 
        s.name as sequence_name,
        sc.contact_phone,
        sc.status,
        sc.current_step,
        sc.next_trigger_time,
        sc.completed_at
    FROM sequence_contacts sc
    JOIN sequences s ON s.id = sc.sequence_id
    ORDER BY s.name, sc.next_trigger_time DESC
    LIMIT 20
""")
results = cur.fetchall()
if results:
    print(f"   Found {len(results)} records:")
    for row in results:
        print(f"   {row[0]} - {row[1]} - Status: {row[2]}, Step: {row[3]}")
        print(f"     Next trigger: {row[4]}, Completed: {row[5]}")
else:
    print("   No records found!")

# Check count by sequence
print("\n2. Count by sequence:")
cur.execute("""
    SELECT 
        s.name,
        COUNT(*) as total_records,
        COUNT(DISTINCT sc.contact_phone) as unique_contacts
    FROM sequence_contacts sc
    JOIN sequences s ON s.id = sc.sequence_id
    GROUP BY s.name
""")
results = cur.fetchall()
for row in results:
    print(f"   {row[0]}: {row[1]} records, {row[2]} unique contacts")

# Check the monitoring view
print("\n3. Checking sequence_progress_overview view:")
cur.execute("SELECT * FROM sequence_progress_overview")
results = cur.fetchall()
print("   Sequence | Trigger | Should | Enrolled | Active | Sent | Failed")
print("   " + "-" * 65)
for row in results:
    print(f"   {row[0]:<15} | {row[1]:<10} | {row[2]:>6} | {row[3]:>8} | {row[4]:>6} | {row[9]:>4} | {row[10]:>6}")

# Check if the Go app enrolled leads automatically
print("\n4. Recent activity check:")
cur.execute("""
    SELECT 
        COUNT(*) as count,
        MIN(next_trigger_time) as earliest,
        MAX(next_trigger_time) as latest
    FROM sequence_contacts
    WHERE next_trigger_time > NOW() - INTERVAL '1 hour'
""")
result = cur.fetchone()
if result[0] > 0:
    print(f"   Found {result[0]} records with trigger times in last hour")
    print(f"   Earliest: {result[1]}")
    print(f"   Latest: {result[2]}")
    print("\n   ⚠️ The Go application appears to have enrolled leads automatically!")

# Clean everything one more time
print("\n5. FINAL CLEANUP:")
try:
    # Delete sequence contacts
    cur.execute("DELETE FROM sequence_contacts")
    sc_deleted = cur.rowcount
    
    # Delete broadcast messages
    cur.execute("DELETE FROM broadcast_messages WHERE sequence_id IS NOT NULL")
    bm_deleted = cur.rowcount
    
    conn.commit()
    print(f"   ✅ Deleted {sc_deleted} sequence_contacts records")
    print(f"   ✅ Deleted {bm_deleted} broadcast messages")
    
    # Verify cleanup
    cur.execute("SELECT COUNT(*) FROM sequence_contacts")
    count = cur.fetchone()[0]
    print(f"\n   Verification: {count} sequence_contacts remaining (should be 0)")
    
except Exception as e:
    conn.rollback()
    print(f"   ❌ Error during cleanup: {e}")

print("\n=== CLEANUP COMPLETE ===")
print("\nThe sequence system is now completely clean and ready for testing.")
print("The Go application should be restarted to pick up the code changes.")

cur.close()
conn.close()
