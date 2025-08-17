import psycopg2
import sys

sys.stdout.reconfigure(encoding='utf-8')

conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cur = conn.cursor()

print("=== CHECKING UNIQUE CONSTRAINTS ON LEADS TABLE ===\n")

# Check all constraints on leads table
cur.execute("""
    SELECT 
        con.conname AS constraint_name,
        con.contype AS constraint_type,
        pg_get_constraintdef(con.oid) AS definition
    FROM pg_constraint con
    JOIN pg_class rel ON rel.oid = con.conrelid
    WHERE rel.relname = 'leads'
    ORDER BY con.contype;
""")

constraints = cur.fetchall()
print(f"Found {len(constraints)} constraints on leads table:\n")

for constraint in constraints:
    print(f"Name: {constraint[0]}")
    print(f"Type: {constraint[1]} ", end="")
    if constraint[1] == 'u':
        print("(UNIQUE)")
    elif constraint[1] == 'p':
        print("(PRIMARY KEY)")
    elif constraint[1] == 'f':
        print("(FOREIGN KEY)")
    elif constraint[1] == 'c':
        print("(CHECK)")
    else:
        print(f"({constraint[1]})")
    print(f"Definition: {constraint[2]}")
    print("-" * 60)

# Check indexes that might enforce uniqueness
print("\n=== CHECKING INDEXES ON LEADS TABLE ===\n")
cur.execute("""
    SELECT 
        indexname,
        indexdef,
        CASE WHEN indexdef LIKE '%UNIQUE%' THEN 'YES' ELSE 'NO' END as is_unique
    FROM pg_indexes
    WHERE tablename = 'leads'
    ORDER BY indexname;
""")

indexes = cur.fetchall()
for index in indexes:
    print(f"Index: {index[0]}")
    print(f"Unique: {index[2]}")
    print(f"Definition: {index[1]}")
    print("-" * 60)

# Check if we have duplicate prevention at database level
print("\n=== ANALYZING DUPLICATE PREVENTION ===\n")

# Check if there's a unique constraint on user_id + phone
cur.execute("""
    SELECT COUNT(*) 
    FROM pg_constraint 
    WHERE conname = 'leads_user_id_phone_key'
""")

has_user_phone_unique = cur.fetchone()[0] > 0

if has_user_phone_unique:
    print("✅ Found UNIQUE constraint on (user_id, phone)")
    print("   This prevents duplicate phone numbers per user")
else:
    print("❌ No UNIQUE constraint on (user_id, phone)")
    print("   Duplicates are possible for same phone number per user")

cur.close()
conn.close()
