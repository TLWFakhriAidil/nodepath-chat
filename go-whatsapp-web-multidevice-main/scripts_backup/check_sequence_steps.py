import psycopg2

conn = psycopg2.connect(
    "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"
)
cur = conn.cursor()

print("=== CHECKING SEQUENCE STEPS DATA ===\n")

# 1. Check sequence steps
print("1. All sequence steps:")
cur.execute("""
    SELECT 
        s.name as sequence_name,
        ss.id as step_id,
        ss.day_number,
        ss.trigger,
        ss.next_trigger,
        ss.is_entry_point,
        ss.trigger_delay_hours
    FROM sequence_steps ss
    JOIN sequences s ON s.id = ss.sequence_id
    ORDER BY s.name, ss.day_number
""")

steps = cur.fetchall()
for step in steps:
    print(f"{step[0]} - Day {step[2]}: {step[3]} -> {step[4]} (Entry: {step[5]}, Delay: {step[6]}h)")
    print(f"  Step ID: {step[1]}")

# 2. Check WARM sequence specifically
print("\n2. WARM Sequence steps in detail:")
cur.execute("""
    SELECT 
        day_number,
        trigger,
        is_entry_point,
        id
    FROM sequence_steps 
    WHERE sequence_id = (SELECT id FROM sequences WHERE name = 'WARM Sequence')
    ORDER BY day_number
""")

warm_steps = cur.fetchall()
for step in warm_steps:
    print(f"  Day {step[0]}: {step[1]} (Entry: {step[2]}) - ID: {step[3]}")

# 3. Check the problematic sequence_contacts
print("\n3. Current sequence_contacts issues:")
cur.execute("""
    SELECT 
        contact_phone,
        current_step,
        status,
        current_trigger,
        sequence_stepid
    FROM sequence_contacts
    WHERE sequence_id = (SELECT id FROM sequences WHERE name = 'WARM Sequence')
    ORDER BY contact_phone, current_step
""")

contacts = cur.fetchall()
for contact in contacts:
    print(f"Phone {contact[0]} - Step {contact[1]}: {contact[2]} ({contact[3]}) - StepID: {contact[4]}")

print("\n4. THE PROBLEM:")
print("It looks like the sequence steps might not have proper day_number values (1,2,3,4)")
print("Or the enrollment is using wrong step IDs")

cur.close()
conn.close()
