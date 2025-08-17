import psycopg2
from datetime import datetime

# Connect to database
conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cur = conn.cursor()

# First, let's check the campaign table structure
print("=== CHECKING CAMPAIGN STRUCTURE ===")
cur.execute("""
    SELECT column_name, data_type 
    FROM information_schema.columns 
    WHERE table_name = 'campaigns' 
    ORDER BY ordinal_position
""")

columns = cur.fetchall()
print("Campaign columns:")
for col in columns:
    print(f"  {col[0]}: {col[1]}")

# Now check campaigns with proper column understanding
print("\n=== CHECKING PENDING CAMPAIGNS ===")
cur.execute("""
    SELECT * FROM campaigns 
    WHERE status = 'pending'
    ORDER BY created_at DESC
    LIMIT 5
""")

# Get column names
col_names = [desc[0] for desc in cur.description]
campaigns = cur.fetchall()

print(f"Found {len(campaigns)} pending campaigns\n")

for row in campaigns:
    campaign = dict(zip(col_names, row))
    print(f"Campaign ID: {campaign['id']}")
    print(f"Title: {campaign['title']}")
    print(f"Status: {campaign['status']}")
    print(f"User ID: {campaign['user_id']}")
    print(f"Device ID: {campaign.get('device_id', 'None')}")
    print(f"Niche: {campaign.get('niche', 'None')}")
    print(f"Target Status: {campaign.get('target_status', 'None')}")
    print(f"Campaign Date: {campaign.get('campaign_date', 'None')}")
    print(f"Time Schedule: {campaign.get('time_schedule', 'None')}")
    print(f"Scheduled At: {campaign.get('scheduled_at', 'None')}")
    print(f"Created At: {campaign['created_at']}")
    
    # Check if it should trigger
    scheduled_at = campaign.get('scheduled_at')
    if scheduled_at:
        print(f"Should trigger? {scheduled_at <= datetime.now()}")
    elif campaign.get('campaign_date') and campaign.get('time_schedule'):
        print(f"Using legacy date/time fields")
    
    # Check for matching leads
    niche = campaign.get('niche')
    target_status = campaign.get('target_status', 'prospect')
    
    if campaign['user_id'] and niche:
        # Fix the query - target_status might be swapped
        cur.execute("""
            SELECT COUNT(*), COUNT(DISTINCT device_id)
            FROM leads 
            WHERE user_id = %s 
            AND niche = %s 
            AND device_id IS NOT NULL
        """, (campaign['user_id'], niche))
        
        result = cur.fetchone()
        print(f"Leads with niche '{niche}': {result[0]} (across {result[1]} devices)")
        
        # Also check with status filter
        if target_status and target_status not in ['all', 'None']:
            cur.execute("""
                SELECT COUNT(*) 
                FROM leads 
                WHERE user_id = %s 
                AND niche = %s 
                AND status = %s
                AND device_id IS NOT NULL
            """, (campaign['user_id'], niche, target_status))
            
            status_count = cur.fetchone()[0]
            print(f"Leads with status '{target_status}': {status_count}")
    
    print("-" * 60)

# Check if any campaigns were processed recently
print("\n=== RECENTLY PROCESSED CAMPAIGNS ===")
cur.execute("""
    SELECT id, title, status, updated_at
    FROM campaigns 
    WHERE updated_at > NOW() - INTERVAL '24 hours'
    AND status != 'pending'
    ORDER BY updated_at DESC
    LIMIT 10
""")

for row in cur.fetchall():
    print(f"ID: {row[0]}, Title: {row[1]}, Status: {row[2]}, Updated: {row[3]}")

# Check broadcast messages from campaigns
print("\n=== CAMPAIGN BROADCAST MESSAGES ===")
cur.execute("""
    SELECT campaign_id, COUNT(*) as count, MAX(created_at) as latest
    FROM broadcast_messages 
    WHERE campaign_id IS NOT NULL
    GROUP BY campaign_id
    ORDER BY latest DESC
    LIMIT 10
""")

messages = cur.fetchall()
if messages:
    for row in messages:
        print(f"Campaign {row[0]}: {row[1]} messages, Latest: {row[2]}")
else:
    print("No broadcast messages from campaigns found")

cur.close()
conn.close()
