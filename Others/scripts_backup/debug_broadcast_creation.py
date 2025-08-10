import psycopg2
import sys
from datetime import datetime, timedelta

sys.stdout.reconfigure(encoding='utf-8')

conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cur = conn.cursor()

# Check if campaign 68 has any broadcast messages
print("=== CHECKING BROADCAST MESSAGES FOR CAMPAIGN 68 ===")
cur.execute("""
    SELECT id, recipient_phone, status, scheduled_at, created_at, error_message
    FROM broadcast_messages 
    WHERE campaign_id = 68
    ORDER BY created_at DESC
""")

messages = cur.fetchall()
print(f"Found {len(messages)} broadcast messages for campaign 68\n")

if messages:
    for msg in messages:
        print(f"Message ID: {msg[0]}")
        print(f"Phone: {msg[1]}")
        print(f"Status: {msg[2]}")
        print(f"Scheduled: {msg[3]}")
        print(f"Created: {msg[4]}")
        print(f"Error: {msg[5]}")
        print("-" * 50)
else:
    print("❌ NO BROADCAST MESSAGES CREATED!")

# Check campaign details
print("\n=== CAMPAIGN 68 DETAILS ===")
cur.execute("""
    SELECT id, title, status, niche, target_status, user_id, device_id,
           campaign_date, time_schedule, scheduled_at, created_at, updated_at
    FROM campaigns 
    WHERE id = 68
""")

campaign = cur.fetchone()
if campaign:
    print(f"ID: {campaign[0]}")
    print(f"Title: {campaign[1]}")
    print(f"Status: {campaign[2]}")
    print(f"Niche: {campaign[3]}")
    print(f"Target Status: {campaign[4]}")
    print(f"User ID: {campaign[5]}")
    print(f"Device ID: {campaign[6]}")
    print(f"Campaign Date: {campaign[7]}")
    print(f"Time Schedule: {campaign[8]}")
    print(f"Scheduled At: {campaign[9]}")
    print(f"Created: {campaign[10]}")
    print(f"Updated: {campaign[11]}")

# Check if the campaign trigger query would find this campaign
print("\n=== CHECKING IF CAMPAIGN SHOULD TRIGGER ===")
cur.execute("""
    SELECT 
        c.id, c.title, c.status,
        CASE 
            WHEN c.scheduled_at IS NOT NULL AND c.scheduled_at <= CURRENT_TIMESTAMP THEN 'YES - scheduled_at passed'
            WHEN c.scheduled_at IS NULL AND 
                 (c.campaign_date || ' ' || COALESCE(c.time_schedule, '00:00:00'))::TIMESTAMP AT TIME ZONE 'Asia/Kuala_Lumpur' <= CURRENT_TIMESTAMP THEN 'YES - date/time passed'
            ELSE 'NO - not time yet'
        END as should_trigger,
        c.scheduled_at,
        (c.campaign_date || ' ' || COALESCE(c.time_schedule, '00:00:00'))::TIMESTAMP AT TIME ZONE 'Asia/Kuala_Lumpur' as calculated_time,
        CURRENT_TIMESTAMP as now
    FROM campaigns c
    WHERE c.id = 68
""")

result = cur.fetchone()
if result:
    print(f"Campaign: {result[1]}")
    print(f"Should Trigger: {result[3]}")
    print(f"Scheduled At: {result[4]}")
    print(f"Calculated Time: {result[5]}")
    print(f"Current Time: {result[6]}")

# Check logs for any errors
print("\n=== CHECKING FOR RECENT ERRORS ===")
cur.execute("""
    SELECT COUNT(*) 
    FROM broadcast_messages 
    WHERE created_at > NOW() - INTERVAL '1 hour'
    AND campaign_id IS NOT NULL
""")

recent_count = cur.fetchone()[0]
print(f"Broadcast messages created in last hour: {recent_count}")

# Check if trigger processor is finding campaigns
print("\n=== SIMULATING CAMPAIGN TRIGGER QUERY ===")
cur.execute("""
    SELECT 
        c.id, c.user_id, c.title, c.message, c.niche, 
        COALESCE(c.target_status, 'all') as target_status, 
        COALESCE(c.image_url, '') as image_url, c.min_delay_seconds, c.max_delay_seconds,
        c.campaign_date, c.time_schedule
    FROM campaigns c
    WHERE c.status = 'pending'
    AND (
        (c.scheduled_at IS NOT NULL AND c.scheduled_at <= CURRENT_TIMESTAMP)
        OR
        (c.scheduled_at IS NULL AND 
         (c.campaign_date || ' ' || COALESCE(c.time_schedule, '00:00:00'))::TIMESTAMP AT TIME ZONE 'Asia/Kuala_Lumpur' <= CURRENT_TIMESTAMP)
    )
""")

trigger_campaigns = cur.fetchall()
print(f"Campaigns that should trigger: {len(trigger_campaigns)}")
for tc in trigger_campaigns:
    print(f"  - Campaign {tc[0]}: {tc[2]}")

cur.close()
conn.close()
