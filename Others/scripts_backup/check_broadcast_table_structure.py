import psycopg2
import sys

sys.stdout.reconfigure(encoding='utf-8')

conn = psycopg2.connect(
    "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"
)
cursor = conn.cursor()

print("=== BROADCAST_MESSAGES TABLE STRUCTURE ===\n")

cursor.execute("""
    SELECT column_name, data_type, is_nullable
    FROM information_schema.columns
    WHERE table_name = 'broadcast_messages'
    ORDER BY ordinal_position
""")

columns = cursor.fetchall()
for col in columns:
    print(f"{col[0]:<20} {col[1]:<20} Nullable: {col[2]}")

print("\n=== COMPARING WITH CAMPAIGN STRUCTURE ===\n")

# Let's see how the actual campaign trigger creates messages
cursor.execute("""
    SELECT 
        bm.id,
        bm.campaign_id,
        bm.recipient_phone,
        bm.status,
        bm.created_at,
        c.min_delay_seconds,
        c.max_delay_seconds
    FROM broadcast_messages bm
    JOIN campaigns c ON c.id = bm.campaign_id
    WHERE bm.campaign_id = 59
""")

result = cursor.fetchone()
if result:
    print("Campaign 59 broadcast message:")
    print(f"  Message ID: {result[0]}")
    print(f"  Status: {result[3]}")
    print(f"  Created: {result[4]}")
    print(f"  Campaign delays: {result[5]}-{result[6]} seconds")
    print("\nNOTE: Delays are stored in CAMPAIGNS table, not broadcast_messages!")

print("\n=== THE ISSUE ===")
print("The min/max delay seconds are stored in the CAMPAIGNS table,")
print("NOT in the broadcast_messages table!")
print("\nThis explains why the campaign trigger is failing -")
print("it's trying to insert into columns that don't exist.")

conn.close()
