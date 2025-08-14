import psycopg2
import sys
from datetime import datetime

sys.stdout.reconfigure(encoding='utf-8')

conn = psycopg2.connect(
    "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"
)
cursor = conn.cursor()

print("=== FINAL CAMPAIGN STATUS CHECK ===\n")

# Check all recent campaigns
cursor.execute("""
    SELECT 
        c.id, c.title, c.status, c.niche, c.target_status,
        c.campaign_date, c.time_schedule,
        c.min_delay_seconds, c.max_delay_seconds,
        COUNT(bm.id) as message_count,
        SUM(CASE WHEN bm.status = 'sent' THEN 1 ELSE 0 END) as sent_count,
        SUM(CASE WHEN bm.status = 'failed' THEN 1 ELSE 0 END) as failed_count
    FROM campaigns c
    LEFT JOIN broadcast_messages bm ON c.id = bm.campaign_id
    WHERE c.created_at > NOW() - INTERVAL '24 hours'
    GROUP BY c.id, c.title, c.status, c.niche, c.target_status,
             c.campaign_date, c.time_schedule, c.min_delay_seconds, c.max_delay_seconds
    ORDER BY c.id DESC
    LIMIT 5
""")

campaigns = cursor.fetchall()

print("Recent Campaigns (Last 24 Hours):")
print("-" * 80)

for campaign in campaigns:
    print(f"\nCampaign {campaign[0]}: {campaign[1]}")
    print(f"  Status: {campaign[2]}")
    print(f"  Niche: {campaign[3]}, Target: {campaign[4]}")
    print(f"  Schedule: {campaign[5]} {campaign[6]}")
    print(f"  Delays: {campaign[7]}-{campaign[8]} seconds")
    print(f"  Messages: {campaign[9]} total ({campaign[10]} sent, {campaign[11]} failed)")

# Check if broadcast processor is working
print("\n\n=== BROADCAST PROCESSOR STATUS ===")
cursor.execute("""
    SELECT 
        COUNT(*) as total_pending,
        MIN(created_at) as oldest_pending,
        MAX(created_at) as newest_pending
    FROM broadcast_messages
    WHERE status = 'pending'
    AND scheduled_at <= NOW()
""")

pending = cursor.fetchone()
if pending[0] > 0:
    print(f"⚠️ {pending[0]} messages pending to be processed!")
    print(f"   Oldest: {pending[1]}")
    print(f"   Newest: {pending[2]}")
else:
    print("✅ No pending messages - broadcast processor is up to date")

# Check device status
print("\n=== DEVICE STATUS ===")
cursor.execute("""
    SELECT id, device_name, status, platform, last_seen
    FROM user_devices
    WHERE user_id = 'de078f16-3266-4ab3-8153-a248b015228f'
    ORDER BY last_seen DESC
""")

for device in cursor.fetchall():
    print(f"\nDevice: {device[1]}")
    print(f"  ID: {device[0]}")
    print(f"  Status: {device[2]}")
    print(f"  Platform: {device[3] or 'WhatsApp Web'}")
    print(f"  Last seen: {device[4]}")

conn.close()

print("\n=== SUMMARY ===")
print("✅ Campaign system fixes applied:")
print("   1. Fixed MinDelay/MaxDelay insert issue")
print("   2. Fixed NULL image_url scan error")
print("   3. Campaigns now work like sequences")
print("\nIf campaigns still aren't creating messages, check:")
print("   - Device is online")
print("   - Leads exist with matching niche/status")
print("   - Campaign time hasn't already passed")
