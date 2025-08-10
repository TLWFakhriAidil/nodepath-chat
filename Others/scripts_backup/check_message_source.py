import psycopg2
import sys
import io
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')

# Connect to database
conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cursor = conn.cursor()

phone = '60138847465'

print("="*100)
print(f"CHECKING MESSAGE SOURCE FOR PHONE: {phone}")
print("="*100)

# Get all messages with clear identification of source
cursor.execute("""
    SELECT 
        bm.id,
        bm.status,
        bm.scheduled_at,
        bm.campaign_id,
        bm.sequence_id,
        CASE 
            WHEN bm.campaign_id IS NOT NULL THEN 'CAMPAIGN'
            WHEN bm.sequence_id IS NOT NULL THEN 'SEQUENCE'
            ELSE 'UNKNOWN'
        END as message_source,
        c.title as campaign_name,
        s.name as sequence_name,
        bm.content
    FROM broadcast_messages bm
    LEFT JOIN campaigns c ON c.id = bm.campaign_id
    LEFT JOIN sequences s ON s.id = bm.sequence_id
    WHERE bm.recipient_phone = %s
    ORDER BY bm.scheduled_at ASC
""", (phone,))

messages = cursor.fetchall()

print(f"\nTotal messages found: {len(messages)}")
print("\n" + "-"*100)
print(f"{'Message ID':<40} {'Source':<10} {'Status':<10} {'Campaign/Sequence Name':<40}")
print("-"*100)

campaign_count = 0
sequence_count = 0

for msg in messages:
    msg_id, status, scheduled, camp_id, seq_id, source, camp_name, seq_name, content = msg
    
    if source == 'CAMPAIGN':
        campaign_count += 1
        name = camp_name or f"Campaign ID: {camp_id}"
    elif source == 'SEQUENCE':
        sequence_count += 1
        name = seq_name or f"Sequence ID: {seq_id}"
    else:
        name = "N/A"
    
    print(f"{msg_id:<40} {source:<10} {status:<10} {name:<40}")

print("\n" + "="*100)
print("SUMMARY:")
print(f"  - Total Messages: {len(messages)}")
print(f"  - From CAMPAIGNS: {campaign_count}")
print(f"  - From SEQUENCES: {sequence_count}")
print("="*100)

# Show detailed breakdown
print("\n\nDETAILED BREAKDOWN:")
print("-"*100)

# Check campaigns
cursor.execute("""
    SELECT 
        c.id,
        c.title,
        COUNT(bm.id) as message_count,
        MIN(bm.scheduled_at) as first_scheduled,
        MAX(bm.scheduled_at) as last_scheduled
    FROM broadcast_messages bm
    JOIN campaigns c ON c.id = bm.campaign_id
    WHERE bm.recipient_phone = %s
    GROUP BY c.id, c.title
""", (phone,))

campaigns = cursor.fetchall()
if campaigns:
    print("\n📢 CAMPAIGNS:")
    for camp in campaigns:
        print(f"\n  Campaign ID: {camp[0]}")
        print(f"  Title: {camp[1]}")
        print(f"  Messages: {camp[2]}")
        print(f"  First scheduled: {camp[3]}")
        print(f"  Last scheduled: {camp[4]}")
else:
    print("\n📢 CAMPAIGNS: None")

# Check sequences
cursor.execute("""
    SELECT 
        s.id,
        s.name,
        s.trigger,
        COUNT(bm.id) as message_count,
        MIN(bm.scheduled_at) as first_scheduled,
        MAX(bm.scheduled_at) as last_scheduled
    FROM broadcast_messages bm
    JOIN sequences s ON s.id = bm.sequence_id
    WHERE bm.recipient_phone = %s
    GROUP BY s.id, s.name, s.trigger
""", (phone,))

sequences = cursor.fetchall()
if sequences:
    print("\n\n🔄 SEQUENCES:")
    for seq in sequences:
        print(f"\n  Sequence ID: {seq[0]}")
        print(f"  Name: {seq[1]}")
        print(f"  Trigger: {seq[2] or 'No trigger set'}")
        print(f"  Messages: {seq[3]}")
        print(f"  First scheduled: {seq[4]}")
        print(f"  Last scheduled: {seq[5]}")
else:
    print("\n\n🔄 SEQUENCES: None")

conn.close()
print("\n" + "="*100)
print("Analysis complete!")
