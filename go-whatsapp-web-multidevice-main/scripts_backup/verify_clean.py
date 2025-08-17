import psycopg2

conn = psycopg2.connect(
    "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"
)
cur = conn.cursor()

print("=== FINAL VERIFICATION ===")

# Check sequence_contacts
cur.execute("SELECT COUNT(*) FROM sequence_contacts")
sc_count = cur.fetchone()[0]
print(f"\nsequence_contacts table: {sc_count} records")

# Check broadcast_messages
cur.execute("SELECT COUNT(*) FROM broadcast_messages WHERE sequence_id IS NOT NULL")
bm_count = cur.fetchone()[0]
print(f"broadcast_messages with sequence_id: {bm_count} records")

# Check the monitoring view
print("\nSequence Progress Overview:")
cur.execute("SELECT * FROM sequence_progress_overview")
results = cur.fetchall()
print("Sequence Name    | Should | Enrolled | Active | Sent | Failed")
print("-" * 65)
for r in results:
    print(f"{r[0]:<16} | {r[2]:>6} | {r[3]:>8} | {r[4]:>6} | {r[9]:>4} | {r[10]:>6}")

# Success message
if sc_count == 0 and bm_count == 0:
    print("\nSUCCESS: All sequence data has been completely deleted!")
    print("The system is now clean and ready for testing.")
else:
    print(f"\nWARNING: Found {sc_count + bm_count} remaining records!")

cur.close()
conn.close()
