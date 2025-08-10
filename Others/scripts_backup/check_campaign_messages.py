import psycopg2
from datetime import datetime, timedelta

# Database connection
conn = psycopg2.connect(
    "postgresql://postgres:postgres@localhost:5432/whatsapp"
)

cursor = conn.cursor()

# Check campaigns 59 and 60
cursor.execute("""
    SELECT id, title, status, niche, target_status, 
           campaign_date, time_schedule, scheduled_at,
           created_at, updated_at
    FROM campaigns
    WHERE id IN (59, 60)
""")
campaigns = cursor.fetchall()

print("Campaigns:")
for campaign in campaigns:
    print(f"ID: {campaign[0]}, Title: {campaign[1]}, Status: {campaign[2]}")
    print(f"  Niche: {campaign[3]}, Target Status: {campaign[4]}")
    print(f"  Campaign Date: {campaign[5]}, Time: {campaign[6]}")
    print(f"  Scheduled At: {campaign[7]}")
    print(f"  Created: {campaign[8]}, Updated: {campaign[9]}")
    print()

# Check if any broadcast messages exist for these campaigns
cursor.execute("""
    SELECT campaign_id, count(*), 
           count(case when status = 'pending' then 1 end) as pending,
           count(case when status = 'sent' then 1 end) as sent,
           count(case when status = 'failed' then 1 end) as failed,
           min(created_at) as first_created,
           max(created_at) as last_created
    FROM broadcast_messages
    WHERE campaign_id IN (59, 60)
    GROUP BY campaign_id
""")
messages = cursor.fetchall()

print("\nBroadcast Messages:")
for msg in messages:
    print(f"Campaign {msg[0]}: Total={msg[1]}, Pending={msg[2]}, Sent={msg[3]}, Failed={msg[4]}")
    print(f"  First created: {msg[5]}, Last created: {msg[6]}")

# Check last 10 broadcast messages with errors
cursor.execute("""
    SELECT id, campaign_id, status, error_message, created_at
    FROM broadcast_messages
    WHERE campaign_id IN (59, 60) AND error_message IS NOT NULL
    ORDER BY created_at DESC
    LIMIT 10
""")
errors = cursor.fetchall()

if errors:
    print("\nRecent Errors:")
    for err in errors:
        print(f"ID: {err[0]}, Campaign: {err[1]}, Status: {err[2]}")
        print(f"  Error: {err[3]}")
        print(f"  Created: {err[4]}")

# Check the lead that's being found
cursor.execute("""
    SELECT phone, name, device_id, user_id, niche, target_status
    FROM leads
    WHERE phone = '60108924904'
""")
lead = cursor.fetchone()

print("\nLead Details:")
if lead:
    print(f"Phone: {lead[0]}, Name: {lead[1]}")
    print(f"Device ID: {lead[2]}")
    print(f"User ID: {lead[3]}")
    print(f"Niche: {lead[4]}, Target Status: {lead[5]}")

conn.close()
