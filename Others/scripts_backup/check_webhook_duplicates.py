import psycopg2
import sys

sys.stdout.reconfigure(encoding='utf-8')

conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cur = conn.cursor()

print("=== CHECKING LEADS TABLE CONSTRAINTS ===\n")

# Check table constraints
cur.execute("""
    SELECT 
        tc.constraint_name,
        tc.constraint_type,
        kcu.column_name,
        tc.table_name
    FROM information_schema.table_constraints tc
    JOIN information_schema.key_column_usage kcu 
        ON tc.constraint_name = kcu.constraint_name 
        AND tc.table_schema = kcu.table_schema
    WHERE tc.table_name = 'leads'
    AND tc.constraint_type IN ('PRIMARY KEY', 'UNIQUE')
    ORDER BY tc.constraint_type, tc.constraint_name
""")

constraints = cur.fetchall()
print("Current constraints on leads table:")
for constraint in constraints:
    print(f"  {constraint[0]}: {constraint[1]} on column(s) {constraint[2]}")

# Check indexes
print("\n=== CHECKING INDEXES ===")
cur.execute("""
    SELECT 
        indexname,
        indexdef
    FROM pg_indexes
    WHERE tablename = 'leads'
    AND indexname LIKE '%unique%' OR indexname LIKE '%phone%'
""")

indexes = cur.fetchall()
if indexes:
    print("Unique/Phone-related indexes:")
    for idx in indexes:
        print(f"  {idx[0]}: {idx[1]}")
else:
    print("No unique or phone-related indexes found")

# Check recent webhook activity
print("\n=== CHECKING RECENT LEAD CREATION PATTERNS ===")
cur.execute("""
    SELECT 
        phone,
        device_id,
        niche,
        COUNT(*) as count,
        array_agg(DISTINCT name) as names,
        MIN(created_at) as first_created,
        MAX(created_at) as last_created,
        MAX(created_at) - MIN(created_at) as time_span
    FROM leads
    WHERE created_at > NOW() - INTERVAL '7 days'
    GROUP BY phone, device_id, niche
    HAVING COUNT(*) > 1
    ORDER BY COUNT(*) DESC
    LIMIT 20
""")

patterns = cur.fetchall()
if patterns:
    print(f"Found {len(patterns)} potential duplicate patterns in last 7 days:\n")
    for p in patterns:
        print(f"Phone: {p[0]}, Device: {p[1][:8]}..., Niche: {p[2]}")
        print(f"  Count: {p[3]}, Names: {p[4]}")
        print(f"  First: {p[5]}, Last: {p[6]}")
        print(f"  Time span: {p[7]}")
        print()
else:
    print("No duplicate patterns found in recent data")

# Check webhook logs if available
print("\n=== SIMULATING WEBHOOK DUPLICATE CHECK ===")
test_cases = [
    ("60108924904", "d409cadc-75e2-4004-a789-c2bad0b31393", "GRR"),
    ("601171219823", "423f7eb9-a742-4ded-b683-76e4df9ba79d", "Property"),
]

for phone, device_id, niche in test_cases:
    cur.execute("""
        SELECT COUNT(*) 
        FROM leads 
        WHERE phone = %s AND device_id = %s AND niche = %s
    """, (phone, device_id, niche))
    
    count = cur.fetchone()[0]
    print(f"Phone: {phone}, Device: {device_id[:8]}..., Niche: {niche}")
    print(f"  Existing leads: {count}")

cur.close()
conn.close()
