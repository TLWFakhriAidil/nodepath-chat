import psycopg2
import sys
from datetime import datetime

sys.stdout.reconfigure(encoding='utf-8')

conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cur = conn.cursor()

print("=== CAMPAIGN 69 STATUS ===")
cur.execute("""
    SELECT id, title, status, niche, target_status, created_at, updated_at
    FROM campaigns 
    WHERE id = 69
""")

campaign = cur.fetchone()
if campaign:
    print(f"ID: {campaign[0]}")
    print(f"Title: {campaign[1]}")
    print(f"Status: {campaign[2]}")
    print(f"Niche: {campaign[3]}")
    print(f"Target Status: {campaign[4]}")
    print(f"Created: {campaign[5]}")
    print(f"Updated: {campaign[6]}")

print("\n=== BROADCAST MESSAGES FOR CAMPAIGN 69 ===")
cur.execute("""
    SELECT id, recipient_phone, status, created_at, error_message
    FROM broadcast_messages 
    WHERE campaign_id = 69
    ORDER BY created_at DESC
""")

messages = cur.fetchall()
print(f"Found {len(messages)} broadcast messages")

for msg in messages[:5]:  # Show first 5
    print(f"\nMessage ID: {msg[0]}")
    print(f"Phone: {msg[1]}")
    print(f"Status: {msg[2]}")
    print(f"Created: {msg[3]}")
    if msg[4]:
        print(f"Error: {msg[4]}")

# Check if campaign trigger processor is running
print("\n=== CHECKING CAMPAIGN PROCESSOR LOGS ===")
cur.execute("""
    SELECT id, title, status, 
           CASE 
               WHEN status = 'triggered' THEN 'Campaign was triggered successfully'
               WHEN status = 'finished' THEN 'Campaign finished (no leads found)'
               WHEN status = 'pending' THEN 'Still waiting to be processed'
               ELSE 'Unknown status'
           END as status_meaning
    FROM campaigns 
    WHERE id IN (68, 69)
    ORDER BY id
""")

for row in cur.fetchall():
    print(f"\nCampaign {row[0]} ({row[1]}): {row[2]}")
    print(f"  → {row[3]}")

# Check if there's an issue with the optimized_campaign_trigger
print("\n=== CHECKING RECENT CAMPAIGN UPDATES ===")
cur.execute("""
    SELECT id, title, status, updated_at
    FROM campaigns 
    WHERE updated_at > NOW() - INTERVAL '10 minutes'
    ORDER BY updated_at DESC
""")

recent = cur.fetchall()
if recent:
    print(f"\nRecently updated campaigns:")
    for r in recent:
        print(f"  Campaign {r[0]} ({r[1]}): {r[2]} at {r[3]}")
else:
    print("\nNo campaigns updated in last 10 minutes")

# Check the lead that's being detected
print("\n=== LEAD DETAILS ===")
cur.execute("""
    SELECT phone, name, status, niche, device_id
    FROM leads 
    WHERE phone = '60108924904'
""")

lead = cur.fetchone()
if lead:
    print(f"Phone: {lead[0]}")
    print(f"Name: {lead[1]}")
    print(f"Status: {lead[2]}")
    print(f"Niche: {lead[3]}")
    print(f"Device ID: {lead[4]}")

cur.close()
conn.close()
