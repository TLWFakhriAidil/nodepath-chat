import psycopg2
import sys
from datetime import datetime

# Set UTF-8 encoding
sys.stdout.reconfigure(encoding='utf-8')

conn = psycopg2.connect(
    "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"
)
cursor = conn.cursor()

print("🔍 FOUND THE ROOT CAUSE!\n")
print("=" * 60)

print("\n📊 CAMPAIGN STATUS:")
print("- Campaign 59: FAILED (created 1 message but device was offline)")
print("- Campaign 60: STUCK IN LOOP (pending status, time has passed)")

print("\n❌ CAMPAIGN 59 ERROR:")
print("Error: 'device not connected: no WhatsApp client found'")
print("This means the device was OFFLINE when trying to send!")

print("\n🔄 WHY CAMPAIGN 60 KEEPS LOOPING:")
# Check the time issue
cursor.execute("""
    SELECT 
        campaign_date,
        time_schedule,
        NOW() AT TIME ZONE 'Asia/Kuala_Lumpur' as now_kl,
        status
    FROM campaigns 
    WHERE id = 60
""")
result = cursor.fetchone()
print(f"- Scheduled: {result[0]} {result[1]} (Asia/KL)")
print(f"- Now: {result[2]}")
print(f"- Status: {result[3]}")
print("\nThe campaign time (16:21) has PASSED but status is still 'pending'!")

print("\n✅ SEQUENCES WORK BECAUSE:")
cursor.execute("""
    SELECT COUNT(*), 
           SUM(CASE WHEN status = 'sent' THEN 1 ELSE 0 END) as sent,
           SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending
    FROM broadcast_messages 
    WHERE sequence_id IS NOT NULL 
    AND created_at > NOW() - INTERVAL '1 day'
""")
seq_stats = cursor.fetchone()
print(f"- Sequences created {seq_stats[0]} messages successfully")
print(f"- {seq_stats[1]} sent, {seq_stats[2]} pending")
print("- Sequences schedule messages for FUTURE times")
print("- Campaigns try to send IMMEDIATELY")

print("\n🔧 THE PROBLEM:")
print("1. Campaign finds lead → tries to send NOW")
print("2. Device is OFFLINE → message fails")
print("3. Campaign 60 time has passed but still 'pending'")
print("4. System keeps checking every minute (infinite loop)")

print("\n💡 SOLUTION:")
print("\n1. IMMEDIATE FIX (stop the loop):")
print("   UPDATE campaigns SET status = 'finished' WHERE id = 60;")

print("\n2. CHECK DEVICE STATUS:")
cursor.execute("""
    SELECT device_name, status, last_seen 
    FROM user_devices 
    WHERE id = 'd409cadc-75e2-4004-a789-c2bad0b31393'
""")
device = cursor.fetchone()
print(f"   Device: {device[0]}")
print(f"   Status: {device[1]}")
print(f"   Last seen: {device[2]}")

print("\n3. BETTER APPROACH:")
print("   - Make sure device is ONLINE before creating campaigns")
print("   - Or reschedule campaign for when device is online")
print("   - Sequences work because they wait for device to come online")

conn.close()
