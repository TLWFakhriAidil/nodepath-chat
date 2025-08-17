import psycopg2
from datetime import datetime

# Connect to database
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"
conn = psycopg2.connect(DB_URI)
cur = conn.cursor()

print("=== ANALYZING SEQUENCE CONTACTS ISSUE ===\n")

# 1. Check current state
print("1. Current sequence_contacts:")
cur.execute("""
    SELECT sequence_id, contact_phone, current_step, status, current_trigger
    FROM sequence_contacts
    ORDER BY contact_phone, current_step
""")
for row in cur.fetchall():
    print(f"  Phone: {row[1]}, Step: {row[2]}, Status: {row[3]}, Trigger: {row[4]}")

# 2. Check for duplicates
print("\n2. Checking for duplicate active steps:")
cur.execute("""
    SELECT contact_phone, COUNT(*) as active_count
    FROM sequence_contacts
    WHERE status = 'active'
    GROUP BY sequence_id, contact_phone
    HAVING COUNT(*) > 1
""")
duplicates = cur.fetchall()
if duplicates:
    print("  Found duplicates:")
    for row in duplicates:
        print(f"    Phone {row[0]} has {row[1]} active steps!")
else:
    print("  No duplicate active steps found")

# 3. Check constraints
print("\n3. Current constraints on sequence_contacts:")
cur.execute("""
    SELECT conname, pg_get_constraintdef(oid) 
    FROM pg_constraint 
    WHERE conrelid = 'sequence_contacts'::regclass
""")
for row in cur.fetchall():
    print(f"  {row[0]}: {row[1]}")

cur.close()
conn.close()