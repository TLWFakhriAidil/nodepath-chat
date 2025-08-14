import psycopg2
from datetime import datetime
import pytz

# Connect to database
conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cur = conn.cursor()

# Check for pending campaigns
print("=== CHECKING CAMPAIGNS ===")
cur.execute("""
    SELECT id, title, status, scheduled_at, time_schedule, niche, target_status, 
           campaign_date, created_at, user_id, device_id, message
    FROM campaigns 
    WHERE status = 'pending'
    ORDER BY created_at DESC
    LIMIT 10
""")

campaigns = cur.fetchall()
print(f"Found {len(campaigns)} pending campaigns\n")

for row in campaigns:
    print(f"Campaign ID: {row[0]}")
    print(f"Title: {row[1]}")
    print(f"Status: {row[2]}")
    print(f"Scheduled: {row[3]} {row[4]}")
    print(f"Campaign Date: {row[6]}")
    print(f"Niche: {row[5]}, Target: {row[7]}")
    print(f"User ID: {row[9]}")
    print(f"Device ID: {row[10]}")
    print(f"Message: {row[11][:50]}..." if row[11] else "No message")
    
    # Check if campaign should be triggered
    if row[3] and row[4]:
        scheduled_str = f"{row[3]} {row[4]}"
        try:
            # Parse the scheduled time
            scheduled_time = datetime.strptime(scheduled_str, "%Y-%m-%d %H:%M:%S")
            # Make it timezone aware (Malaysia time)
            kl_tz = pytz.timezone('Asia/Kuala_Lumpur')
            scheduled_time = kl_tz.localize(scheduled_time)
            
            # Get current time
            now = datetime.now(kl_tz)
            
            print(f"Scheduled Time: {scheduled_time}")
            print(f"Current Time: {now}")
            print(f"Should trigger? {scheduled_time <= now}")
        except Exception as e:
            print(f"Error parsing time: {e}")
    
    # Check for matching leads
    if row[9] and row[5]:  # user_id and niche
        cur.execute("""
            SELECT COUNT(*) 
            FROM leads 
            WHERE user_id = %s 
            AND niche = %s 
            AND status = %s
            AND device_id IS NOT NULL
        """, (row[9], row[5], row[7] or 'prospect'))
        
        lead_count = cur.fetchone()[0]
        print(f"Matching leads: {lead_count}")
        
        # Check connected devices
        cur.execute("""
            SELECT COUNT(*) 
            FROM user_devices 
            WHERE user_id = %s 
            AND (status IN ('connected', 'Connected', 'online', 'Online') OR platform IS NOT NULL)
        """, (row[9],))
        
        device_count = cur.fetchone()[0]
        print(f"Connected devices: {device_count}")
    
    print("-" * 50)

# Check broadcast messages
print("\n=== RECENT BROADCAST MESSAGES ===")
cur.execute("""
    SELECT campaign_id, COUNT(*) as msg_count, MIN(created_at) as first_created
    FROM broadcast_messages 
    WHERE campaign_id IS NOT NULL
    GROUP BY campaign_id
    ORDER BY first_created DESC
    LIMIT 10
""")

for row in cur.fetchall():
    print(f"Campaign ID: {row[0]}, Messages: {row[1]}, Created: {row[2]}")

# Check if campaign trigger is actually running
print("\n=== CHECKING CAMPAIGN PROCESSING ===")
cur.execute("""
    SELECT id, title, status, updated_at
    FROM campaigns 
    WHERE status IN ('triggered', 'finished')
    AND updated_at > NOW() - INTERVAL '1 hour'
    ORDER BY updated_at DESC
    LIMIT 5
""")

recent = cur.fetchall()
if recent:
    print(f"Recently processed campaigns:")
    for row in recent:
        print(f"  ID: {row[0]}, Title: {row[1]}, Status: {row[2]}, Updated: {row[3]}")
else:
    print("No campaigns processed in the last hour")

cur.close()
conn.close()
