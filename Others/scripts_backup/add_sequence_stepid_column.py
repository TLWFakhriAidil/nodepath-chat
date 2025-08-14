import psycopg2

# Connect to PostgreSQL
conn = psycopg2.connect(
    "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"
)
cur = conn.cursor()

print("Checking for sequence_stepid column...")

# Check if column exists
cur.execute("""
    SELECT column_name 
    FROM information_schema.columns 
    WHERE table_name = 'broadcast_messages' 
    AND column_name = 'sequence_stepid'
""")
result = cur.fetchone()

if result:
    print("Column 'sequence_stepid' already exists!")
else:
    print("Column 'sequence_stepid' NOT found. Adding it now...")
    
    try:
        # Add the column
        cur.execute("""
            ALTER TABLE broadcast_messages 
            ADD COLUMN sequence_stepid UUID REFERENCES sequence_steps(id)
        """)
        conn.commit()
        print("Successfully added sequence_stepid column!")
        
        # Verify it was added
        cur.execute("""
            SELECT column_name, data_type 
            FROM information_schema.columns 
            WHERE table_name = 'broadcast_messages' 
            AND column_name = 'sequence_stepid'
        """)
        result = cur.fetchone()
        if result:
            print(f"Verified: Column '{result[0]}' with type '{result[1]}' now exists!")
        
    except Exception as e:
        conn.rollback()
        print(f"Error adding column: {e}")

# Show all columns in broadcast_messages
print("\nAll columns in broadcast_messages table:")
cur.execute("""
    SELECT column_name, data_type, is_nullable
    FROM information_schema.columns 
    WHERE table_name = 'broadcast_messages'
    ORDER BY ordinal_position
""")
columns = cur.fetchall()
for col in columns:
    nullable = 'NULL' if col[2] == 'YES' else 'NOT NULL'
    print(f"  - {col[0]} ({col[1]}) {nullable}")

cur.close()
conn.close()
