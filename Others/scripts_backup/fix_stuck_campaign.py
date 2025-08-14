import psycopg2

conn = psycopg2.connect(
    "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"
)
cursor = conn.cursor()

print("Fixing stuck campaign...")

# Stop the infinite loop
cursor.execute("UPDATE campaigns SET status = 'finished' WHERE id = 60")
conn.commit()

print("SUCCESS: Campaign 60 status updated to 'finished'")
print("The infinite loop should stop now!")

# Verify the update
cursor.execute("SELECT id, title, status FROM campaigns WHERE id = 60")
result = cursor.fetchone()
print(f"\nVerification: Campaign {result[0]} '{result[1]}' is now: {result[2]}")

conn.close()
