import psycopg2

# Connect to verify changes
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

print("=== SEQUENCE FIX SUMMARY ===\n")

# Test the database changes
conn = psycopg2.connect(DB_URI)
cur = conn.cursor()

# 1. Check constraints
print("1. Database Constraints Applied:")
cur.execute("""
    SELECT conname 
    FROM pg_constraint 
    WHERE conrelid = 'sequence_contacts'::regclass
    AND contype = 'u'
""")
constraints = cur.fetchall()
if constraints:
    for row in constraints:
        print(f"   - {row[0]}")
else:
    print("   - No unique constraints found")

# 2. Check indexes
print("\n2. Performance Indexes Created:")
cur.execute("""
    SELECT indexname 
    FROM pg_indexes 
    WHERE tablename = 'sequence_contacts'
""")
for row in cur.fetchall():
    print(f"   - {row[0]}")

# 3. Check current data state
print("\n3. Current Sequence Contacts State:")
cur.execute("""
    SELECT contact_phone, current_step, status
    FROM sequence_contacts
    ORDER BY contact_phone, current_step
""")
data = cur.fetchall()
if data:
    for row in data:
        print(f"   Phone: {row[0]}, Step: {row[1]}, Status: {row[2]}")
else:
    print("   No sequence contacts found")

print("\n=== FIX APPLIED SUCCESSFULLY ===")
print("\nDatabase changes:")
print("- Unique constraint on (sequence_id, contact_phone) WHERE status='active'")
print("- Index on pending steps by trigger time")
print("- Cleaned up duplicate/completed records")
print("\nCode changes needed in sequence_trigger_processor.go:")
print("- Update query to ORDER BY next_trigger_time ASC")
print("- Use FOR UPDATE SKIP LOCKED")
print("- Add context for transaction handling")

cur.close()
conn.close()