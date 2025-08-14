import psycopg2
import sys
from datetime import datetime, timedelta

# Set UTF-8 encoding
sys.stdout.reconfigure(encoding='utf-8')

conn = psycopg2.connect(
    "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"
)
cursor = conn.cursor()

print("=== CHECKING GRR NICHE CAMPAIGNS ===\n")

# 1. Get all campaigns with niche GRR
cursor.execute("""
    SELECT id, title, status, campaign_date, time_schedule, 
           target_status, user_id, created_at, updated_at
    FROM campaigns
    WHERE niche = 'GRR'
    ORDER BY created_at DESC
""")
campaigns = cursor.fetchall()

print(f"Found {len(campaigns)} campaigns with niche 'GRR':\n")
for c in campaigns:
    print(f"Campaign ID: {c[0]} - {c[1]}")
    print(f"  Status: {c[2]}")
    print(f"  Date/Time: {c[3]} {c[4]}")
    print(f"  Target Status: {c[5]}")
    print(f"  Created: {c[7]}")
    print()

# 2. Check broadcast messages created by GRR campaigns
print("\n=== BROADCAST MESSAGES FROM GRR CAMPAIGNS ===\n")
cursor.execute("""
    SELECT 
        c.id as campaign_id,
        c.title,
        COUNT(bm.id) as total_messages,
        SUM(CASE WHEN bm.status = 'pending' THEN 1 ELSE 0 END) as pending,
        SUM(CASE WHEN bm.status = 'sent' THEN 1 ELSE 0 END) as sent,
        SUM(CASE WHEN bm.status = 'failed' THEN 1 ELSE 0 END) as failed,
        MIN(bm.created_at) as first_created,
        MAX(bm.created_at) as last_created
    FROM campaigns c
    LEFT JOIN broadcast_messages bm ON c.id = bm.campaign_id
    WHERE c.niche = 'GRR'
    GROUP BY c.id, c.title
    ORDER BY c.created_at DESC
""")
results = cursor.fetchall()

for r in results:
    print(f"Campaign {r[0]}: {r[1]}")
    if r[2] > 0:
        print(f"  ✅ Created {r[2]} broadcast messages")
        print(f"  Status: {r[4]} sent, {r[3]} pending, {r[5]} failed")
        print(f"  First: {r[6]}, Last: {r[7]}")
    else:
        print(f"  ❌ NO broadcast messages created")
    print()

# 3. Check leads with GRR niche
print("\n=== LEADS WITH GRR NICHE ===\n")
cursor.execute("""
    SELECT 
        COUNT(*) as total_leads,
        COUNT(DISTINCT device_id) as devices,
        COUNT(DISTINCT user_id) as users,
        SUM(CASE WHEN target_status = 'prospect' THEN 1 ELSE 0 END) as prospects,
        SUM(CASE WHEN target_status = 'customer' THEN 1 ELSE 0 END) as customers
    FROM leads
    WHERE niche = 'GRR'
""")
lead_stats = cursor.fetchone()

print(f"Total GRR leads: {lead_stats[0]}")
print(f"Across {lead_stats[1]} devices")
print(f"Owned by {lead_stats[2]} users")
print(f"Prospects: {lead_stats[3]}, Customers: {lead_stats[4]}")

# 4. Show sample GRR leads
print("\n=== SAMPLE GRR LEADS ===\n")
cursor.execute("""
    SELECT phone, name, device_id, user_id, target_status, created_at
    FROM leads
    WHERE niche = 'GRR'
    LIMIT 5
""")
sample_leads = cursor.fetchall()

for lead in sample_leads:
    print(f"Phone: {lead[0]}")
    print(f"  Name: {lead[1]}")
    print(f"  Device: {lead[2]}")
    print(f"  User: {lead[3]}")
    print(f"  Status: {lead[4]}")
    print(f"  Created: {lead[5]}")
    print()

# 5. Check if there's a working GRR campaign
print("\n=== MOST RECENT SUCCESSFUL GRR CAMPAIGN ===\n")
cursor.execute("""
    SELECT 
        c.id, c.title, c.status,
        COUNT(bm.id) as messages_sent
    FROM campaigns c
    JOIN broadcast_messages bm ON c.id = bm.campaign_id
    WHERE c.niche = 'GRR'
    AND bm.status = 'sent'
    GROUP BY c.id, c.title, c.status
    ORDER BY MAX(bm.created_at) DESC
    LIMIT 1
""")
success = cursor.fetchone()

if success:
    print(f"✅ Campaign {success[0]}: {success[1]}")
    print(f"   Successfully sent {success[3]} messages")
else:
    print("❌ No GRR campaigns have successfully sent messages")

# 6. Check devices for GRR leads
print("\n=== DEVICES WITH GRR LEADS ===\n")
cursor.execute("""
    SELECT 
        ud.id, ud.device_name, ud.status,
        COUNT(l.id) as grr_leads
    FROM user_devices ud
    JOIN leads l ON ud.id = l.device_id
    WHERE l.niche = 'GRR'
    GROUP BY ud.id, ud.device_name, ud.status
    ORDER BY grr_leads DESC
""")
devices = cursor.fetchall()

for d in devices:
    print(f"Device: {d[1]} ({d[0]})")
    print(f"  Status: {d[2]}")
    print(f"  GRR Leads: {d[3]}")
    print()

conn.close()

print("\n=== SUMMARY ===")
print("This shows whether GRR campaigns are successfully creating broadcast messages from leads.")
