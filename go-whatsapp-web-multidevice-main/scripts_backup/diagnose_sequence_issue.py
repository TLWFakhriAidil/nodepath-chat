import psycopg2

DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"
conn = psycopg2.connect(DB_URI)
cur = conn.cursor()

print("=== CHECKING CURRENT SEQUENCE CONTACTS ===\n")

# 1. Check all records with their trigger times
print("1. All sequence contacts with trigger times:")
cur.execute("""
    SELECT 
        contact_phone, 
        current_step, 
        status, 
        current_trigger,
        next_trigger_time,
        sequence_stepid
    FROM sequence_contacts
    ORDER BY contact_phone, current_step
""")

for row in cur.fetchall():
    print(f"Phone: {row[0]}, Step: {row[1]}, Status: {row[2]}, Trigger: {row[3]}, Time: {row[4]}, StepID: {row[5]}")

# 2. Check for constraint violations
print("\n2. Checking for multiple active steps per contact:")
cur.execute("""
    SELECT contact_phone, COUNT(*) as active_count
    FROM sequence_contacts
    WHERE status = 'active'
    GROUP BY sequence_id, contact_phone
    HAVING COUNT(*) > 1
""")

violations = cur.fetchall()
if violations:
    print("CONSTRAINT VIOLATION FOUND!")
    for v in violations:
        print(f"  Phone {v[0]} has {v[1]} active steps!")
else:
    print("  No violations (good)")

# 3. Check missing Step 1
print("\n3. Checking for Step 1 records:")
cur.execute("""
    SELECT contact_phone, status, completed_at
    FROM sequence_contacts
    WHERE current_step = 1
""")

step1 = cur.fetchall()
if not step1:
    print("  WARNING: No Step 1 records found!")
else:
    for s in step1:
        print(f"  Phone {s[0]}: status={s[1]}, completed={s[2]}")

# 4. Check the sequence of steps
print("\n4. Step progression analysis:")
cur.execute("""
    SELECT DISTINCT contact_phone
    FROM sequence_contacts
""")

for phone_row in cur.fetchall():
    phone = phone_row[0]
    print(f"\n  Phone {phone}:")
    cur.execute("""
        SELECT current_step, status, next_trigger_time
        FROM sequence_contacts
        WHERE contact_phone = %s
        ORDER BY current_step
    """, (phone,))
    
    for step in cur.fetchall():
        print(f"    Step {step[0]}: {step[1]} (next trigger: {step[2]})")

cur.close()
conn.close()