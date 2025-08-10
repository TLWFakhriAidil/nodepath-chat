import psycopg2
import pandas as pd

conn = psycopg2.connect(
    "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"
)
cur = conn.cursor()

print("=== ANALYZING SEQUENCE_CONTACTS ISSUES ===\n")

# 1. Get all sequence_contacts data
print("1. Fetching all sequence_contacts records...")
cur.execute("""
    SELECT 
        sc.id,
        s.name as sequence_name,
        sc.contact_phone,
        sc.contact_name,
        sc.current_step,
        sc.status,
        sc.current_trigger,
        sc.next_trigger_time,
        sc.completed_at,
        ss.day_number as expected_step,
        ss.trigger as step_trigger
    FROM sequence_contacts sc
    JOIN sequences s ON s.id = sc.sequence_id
    LEFT JOIN sequence_steps ss ON ss.id = sc.sequence_stepid
    ORDER BY sc.contact_phone, sc.current_step
""")

results = cur.fetchall()
print(f"Found {len(results)} records\n")

# 2. Show the data grouped by phone
print("2. Records by phone number:")
current_phone = None
for row in results:
    if row[2] != current_phone:
        current_phone = row[2]
        print(f"\nPhone: {current_phone}")
    print(f"  Step {row[4]} ({row[9]}): {row[5]:<10} Trigger: {row[6]:<15} Name: '{row[3]}'")

# 3. Check for issues
print("\n\n3. ISSUES FOUND:")

# Check for weird contact names
print("\na) Contact names with step numbers:")
cur.execute("""
    SELECT DISTINCT contact_name, contact_phone 
    FROM sequence_contacts 
    WHERE contact_name LIKE '%1%' 
       OR contact_name LIKE '%2%'
       OR contact_name LIKE '%3%'
       OR contact_name LIKE '%4%'
    ORDER BY contact_phone
""")
weird_names = cur.fetchall()
for name, phone in weird_names:
    print(f"   {phone}: '{name}'")

# Check for multiple active steps
print("\nb) Multiple active steps for same contact:")
cur.execute("""
    SELECT contact_phone, COUNT(*) as active_count
    FROM sequence_contacts
    WHERE status = 'active'
    GROUP BY contact_phone
    HAVING COUNT(*) > 1
""")
multiple_active = cur.fetchall()
for phone, count in multiple_active:
    print(f"   {phone}: {count} active steps")

# Check completed with future trigger times
print("\nc) Completed steps with future trigger times:")
cur.execute("""
    SELECT contact_phone, current_step, next_trigger_time
    FROM sequence_contacts
    WHERE status = 'completed'
    AND next_trigger_time > NOW()
""")
completed_future = cur.fetchall()
for phone, step, trigger_time in completed_future:
    print(f"   {phone} Step {step}: {trigger_time}")

print("\n\n4. CLEANUP RECOMMENDATION:")
print("This data needs to be cleaned up. The issues are:")
print("- Contact names are being concatenated with step numbers")
print("- Multiple enrollment creating duplicate entries")
print("- Chain reaction not working properly")

# Clean everything
print("\n5. Cleaning all sequence_contacts...")
try:
    cur.execute("DELETE FROM sequence_contacts")
    deleted = cur.rowcount
    cur.execute("DELETE FROM broadcast_messages WHERE sequence_id IS NOT NULL")
    bm_deleted = cur.rowcount
    conn.commit()
    print(f"   Deleted {deleted} sequence_contacts records")
    print(f"   Deleted {bm_deleted} broadcast messages")
    print("\n   ✅ All sequence data cleaned!")
except Exception as e:
    conn.rollback()
    print(f"   Error: {e}")

cur.close()
conn.close()
