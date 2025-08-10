import psycopg2

# Connect to database
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"
conn = psycopg2.connect(DB_URI)
cur = conn.cursor()

print("Executing sequence contacts fix...")

# Read and execute the SQL file
with open('fix_sequence_contacts.sql', 'r') as f:
    sql = f.read()
    
try:
    cur.execute(sql)
    conn.commit()
    print("Fix applied successfully!")
except Exception as e:
    conn.rollback()
    print(f"Error: {e}")
    
cur.close()
conn.close()