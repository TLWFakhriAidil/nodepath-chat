import psycopg2

DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"
conn = psycopg2.connect(DB_URI)
cur = conn.cursor()

print("=== SEQUENCE_CONTACTS TABLE COLUMNS ===")
cur.execute("""
    SELECT column_name, data_type 
    FROM information_schema.columns 
    WHERE table_name = 'sequence_contacts' 
    ORDER BY ordinal_position
""")
for col in cur.fetchall():
    print(f"{col[0]}: {col[1]}")

print("\n=== CHECKING FOR TRIGGERS ===")
cur.execute("""
    SELECT trigger_name, event_manipulation, action_statement
    FROM information_schema.triggers
    WHERE event_object_table = 'sequence_contacts'
""")
triggers = cur.fetchall()
if triggers:
    for t in triggers:
        print(f"Trigger: {t[0]}, Event: {t[1]}")
        print(f"Action: {t[2][:100]}...")
else:
    print("No triggers found on sequence_contacts table")

cur.close()
conn.close()
