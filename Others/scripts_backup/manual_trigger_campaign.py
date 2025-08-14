import psycopg2
import sys

sys.stdout.reconfigure(encoding='utf-8')

conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cur = conn.cursor()

# First, let's manually trigger campaign 69 by creating a broadcast message
print("=== MANUALLY CREATING BROADCAST MESSAGE FOR CAMPAIGN 69 ===")

# Get campaign details
cur.execute("""
    SELECT c.id, c.user_id, c.device_id, c.message, c.image_url
    FROM campaigns c
    WHERE c.id = 69
""")

campaign = cur.fetchone()
if not campaign:
    print("Campaign 69 not found!")
    exit()

campaign_id = campaign[0]
user_id = campaign[1]
device_id = campaign[2]
message = campaign[3]
image_url = campaign[4]

# Get the lead details
cur.execute("""
    SELECT phone, name, device_id
    FROM leads 
    WHERE niche = 'GRR'
    AND user_id = %s
    LIMIT 1
""", (user_id,))

lead = cur.fetchone()
if not lead:
    print("No GRR lead found!")
    exit()

phone = lead[0]
name = lead[1]
lead_device_id = lead[2]

print(f"Campaign ID: {campaign_id}")
print(f"User ID: {user_id}")
print(f"Lead Phone: {phone}")
print(f"Lead Name: {name}")
print(f"Lead Device ID: {lead_device_id}")
print(f"Message: {message[:50]}...")

# Check if device_id is valid
if not lead_device_id:
    print("\n❌ ERROR: Lead has no device_id!")
    exit()

# Try to create a broadcast message manually
print("\n=== CREATING BROADCAST MESSAGE ===")
try:
    import uuid
    from datetime import datetime
    
    msg_id = str(uuid.uuid4())
    
    cur.execute("""
        INSERT INTO broadcast_messages 
        (id, user_id, device_id, campaign_id, recipient_phone, recipient_name,
         message_type, content, media_url, status, scheduled_at, created_at)
        VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
    """, (
        msg_id,
        user_id,
        lead_device_id,  # Use the lead's device_id
        campaign_id,
        phone,
        name,
        'text',
        message,
        image_url or None,
        'pending',
        datetime.now(),
        datetime.now()
    ))
    
    conn.commit()
    print("✅ Broadcast message created successfully!")
    print(f"Message ID: {msg_id}")
    
    # Update campaign status
    cur.execute("""
        UPDATE campaigns 
        SET status = 'triggered', updated_at = NOW()
        WHERE id = %s
    """, (campaign_id,))
    
    conn.commit()
    print("✅ Campaign status updated to 'triggered'")
    
except Exception as e:
    print(f"❌ ERROR creating broadcast message: {e}")
    conn.rollback()

# Check if message was created
print("\n=== VERIFYING BROADCAST MESSAGE ===")
cur.execute("""
    SELECT COUNT(*) FROM broadcast_messages WHERE campaign_id = 69
""")
count = cur.fetchone()[0]
print(f"Total broadcast messages for campaign 69: {count}")

cur.close()
conn.close()
