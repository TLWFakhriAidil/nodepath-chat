import psycopg2

DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"
conn = psycopg2.connect(DB_URI)
cur = conn.cursor()

print("=== CHECKING TRIGGER FUNCTION ===")
cur.execute("""
    SELECT routine_definition
    FROM information_schema.routines
    WHERE routine_name = 'check_step_sequence'
""")
result = cur.fetchone()
if result:
    print("Function check_step_sequence():")
    print(result[0])
else:
    print("Function not found")

# Also check for any updated_at references in functions
print("\n=== CHECKING FOR UPDATED_AT IN FUNCTIONS ===")
cur.execute("""
    SELECT routine_name, routine_definition
    FROM information_schema.routines
    WHERE routine_definition LIKE '%sequence_contacts%'
    AND routine_definition LIKE '%updated_at%'
""")
functions = cur.fetchall()
if functions:
    for f in functions:
        print(f"\nFunction: {f[0]}")
        print(f"Contains updated_at reference")
else:
    print("No functions with updated_at reference found")

cur.close()
conn.close()
