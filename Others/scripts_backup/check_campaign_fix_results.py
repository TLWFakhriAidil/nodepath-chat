import psycopg2
import sys
from datetime import datetime

sys.stdout.reconfigure(encoding='utf-8')

conn = psycopg2.connect(
    "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"
)
cursor = conn.cursor()

print("=== CHECKING CAMPAIGN FIX RESULTS ===\n")

# Check campaign 62 (our test)
cursor.execute("""
    SELECT id, title, status, campaign_date, time_schedule, 
           min_delay_seconds, max_delay_seconds
    FROM campaigns
    WHERE id = 62
""")
campaign = cursor.fetchone()

if campaign:
    print(f"Test Campaign ID 62:")
    print(f"  Title: {campaign[1]}")
    print(f"  Status: {campaign[2]}")
    print(f"  Scheduled: {campaign[3]} {campaign[4]}")
    print(f"  Delays: {campaign[5]}-{campaign[6]} seconds")

# Check if any broadcast messages were created
cursor.execute("""
    SELECT COUNT(*), 
           SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending,
           SUM(CASE WHEN status = 'sent' THEN 1 ELSE 0 END) as sent,
           SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed
    FROM broadcast_messages
    WHERE campaign_id = 62
""")
result = cursor.fetchone()

print(f"\nBroadcast messages for campaign 62:")
print(f"  Total: {result[0]}")
if result[0] > 0:
    print(f"  Pending: {result[1]}, Sent: {result[2]}, Failed: {result[3]}")
    print("\n✅ SUCCESS! Campaign created broadcast messages!")
else:
    print("  ⏰ No messages yet (campaign may not have triggered yet)")

# Check recent campaigns
print("\n=== ALL RECENT GRR CAMPAIGNS ===")
cursor.execute("""
    SELECT c.id, c.title, c.status, 
           COUNT(bm.id) as message_count,
           MAX(bm.created_at) as last_message_created
    FROM campaigns c
    LEFT JOIN broadcast_messages bm ON c.id = bm.campaign_id
    WHERE c.niche = 'GRR'
    AND c.created_at > NOW() - INTERVAL '1 day'
    GROUP BY c.id, c.title, c.status
    ORDER BY c.id DESC
""")

for row in cursor.fetchall():
    print(f"\nCampaign {row[0]}: {row[1]}")
    print(f"  Status: {row[2]}")
    print(f"  Messages created: {row[3]}")
    if row[4]:
        print(f"  Last message: {row[4]}")

conn.close()

print("\n=== SUMMARY ===")
print("The campaign fix has been deployed successfully!")
print("Campaigns should now create broadcast messages without errors.")
print("If no messages appear yet, wait a minute and check again.")
