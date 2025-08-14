import psycopg2
import sys
from datetime import datetime

sys.stdout.reconfigure(encoding='utf-8')

conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cur = conn.cursor()

print("=== CHECKING FOR DUPLICATE BROADCAST MESSAGES ===\n")

# First, find all duplicates based on device_id + recipient_phone
cur.execute("""
    SELECT 
        device_id, 
        recipient_phone, 
        COUNT(*) as duplicate_count,
        array_agg(id ORDER BY created_at ASC) as message_ids,
        array_agg(status ORDER BY created_at ASC) as statuses,
        array_agg(campaign_id ORDER BY created_at ASC) as campaign_ids,
        array_agg(sequence_id ORDER BY created_at ASC) as sequence_ids,
        array_agg(created_at ORDER BY created_at ASC) as created_dates,
        array_agg(scheduled_at ORDER BY created_at ASC) as scheduled_dates
    FROM broadcast_messages
    WHERE device_id IS NOT NULL 
    AND recipient_phone IS NOT NULL
    GROUP BY device_id, recipient_phone
    HAVING COUNT(*) > 1
    ORDER BY COUNT(*) DESC
    LIMIT 50
""")

duplicates = cur.fetchall()
print(f"Found {len(duplicates)} duplicate phone+device combinations\n")

if not duplicates:
    print("No duplicates found!")
    cur.close()
    conn.close()
    exit()

# Show detailed analysis
print("=== DUPLICATE ANALYSIS ===")

# Count by status
cur.execute("""
    WITH duplicate_messages AS (
        SELECT 
            device_id,
            recipient_phone,
            id,
            status,
            campaign_id,
            sequence_id,
            ROW_NUMBER() OVER (PARTITION BY device_id, recipient_phone ORDER BY created_at ASC) as rn
        FROM broadcast_messages
        WHERE device_id IS NOT NULL 
        AND recipient_phone IS NOT NULL
    )
    SELECT 
        status,
        COUNT(*) as count
    FROM duplicate_messages 
    WHERE rn > 1
    GROUP BY status
    ORDER BY count DESC
""")

status_breakdown = cur.fetchall()
print("\nDuplicate messages by status:")
for status in status_breakdown:
    print(f"  {status[0]}: {status[1]} messages")

# Show sample duplicates
print("\n=== SAMPLE DUPLICATES (First 10) ===")
for i, dup in enumerate(duplicates[:10]):
    device_id = dup[0]
    phone = dup[1]
    count = dup[2]
    message_ids = dup[3]
    statuses = dup[4]
    campaign_ids = dup[5]
    sequence_ids = dup[6]
    created_dates = dup[7]
    scheduled_dates = dup[8]
    
    print(f"\n{i+1}. Phone: {phone}")
    print(f"   Device ID: {device_id}")
    print(f"   Duplicate Count: {count}")
    print("   Messages:")
    for j in range(min(3, len(message_ids))):
        campaign_info = f"Campaign {campaign_ids[j]}" if campaign_ids[j] else f"Sequence {sequence_ids[j]}"
        print(f"     - ID: {message_ids[j][:8]}..., Status: {statuses[j]}, {campaign_info}")
        print(f"       Created: {created_dates[j]}, Scheduled: {scheduled_dates[j]}")
    if len(message_ids) > 3:
        print(f"     ... and {len(message_ids) - 3} more")

# Check if duplicates are from same campaign/sequence
print("\n=== DUPLICATE SOURCE ANALYSIS ===")
cur.execute("""
    WITH dup_check AS (
        SELECT 
            device_id,
            recipient_phone,
            campaign_id,
            sequence_id,
            COUNT(*) as dup_count
        FROM broadcast_messages
        WHERE device_id IS NOT NULL 
        AND recipient_phone IS NOT NULL
        GROUP BY device_id, recipient_phone, campaign_id, sequence_id
        HAVING COUNT(*) > 1
    )
    SELECT 
        CASE 
            WHEN campaign_id IS NOT NULL THEN 'Campaign'
            WHEN sequence_id IS NOT NULL THEN 'Sequence'
            ELSE 'Unknown'
        END as source_type,
        COUNT(*) as duplicate_groups,
        SUM(dup_count) as total_duplicates
    FROM dup_check
    GROUP BY source_type
""")

source_analysis = cur.fetchall()
for source in source_analysis:
    print(f"{source[0]}: {source[1]} duplicate groups with {source[2]} total messages")

# Count total duplicates to potentially remove
cur.execute("""
    WITH duplicate_messages AS (
        SELECT 
            device_id,
            recipient_phone,
            id,
            status,
            ROW_NUMBER() OVER (PARTITION BY device_id, recipient_phone ORDER BY 
                CASE 
                    WHEN status = 'sent' THEN 1
                    WHEN status = 'pending' THEN 2
                    WHEN status = 'queued' THEN 3
                    WHEN status = 'failed' THEN 4
                    ELSE 5
                END,
                created_at ASC
            ) as rn
        FROM broadcast_messages
        WHERE device_id IS NOT NULL 
        AND recipient_phone IS NOT NULL
    )
    SELECT COUNT(*) 
    FROM duplicate_messages 
    WHERE rn > 1
""")

total_to_delete = cur.fetchone()[0]
print(f"\n=== SUMMARY ===")
print(f"Total duplicate messages that could be deleted: {total_to_delete}")
print("(Strategy: Keep the best status message - sent > pending > queued > failed)")

# Show what would be kept vs deleted
print("\n=== DELETION PREVIEW ===")
cur.execute("""
    WITH duplicate_messages AS (
        SELECT 
            device_id,
            recipient_phone,
            id,
            status,
            campaign_id,
            sequence_id,
            ROW_NUMBER() OVER (PARTITION BY device_id, recipient_phone ORDER BY 
                CASE 
                    WHEN status = 'sent' THEN 1
                    WHEN status = 'pending' THEN 2
                    WHEN status = 'queued' THEN 3
                    WHEN status = 'failed' THEN 4
                    ELSE 5
                END,
                created_at ASC
            ) as rn
        FROM broadcast_messages
        WHERE device_id IS NOT NULL 
        AND recipient_phone IS NOT NULL
    )
    SELECT 
        CASE WHEN rn = 1 THEN 'KEEP' ELSE 'DELETE' END as action,
        status,
        COUNT(*) as count
    FROM duplicate_messages
    GROUP BY action, status
    ORDER BY action, count DESC
""")

preview = cur.fetchall()
keep_messages = []
delete_messages = []

for row in preview:
    if row[0] == 'KEEP':
        keep_messages.append(row)
    else:
        delete_messages.append(row)

print("\nMessages to KEEP:")
for msg in keep_messages:
    print(f"  {msg[1]}: {msg[2]} messages")

print("\nMessages to DELETE:")
for msg in delete_messages:
    print(f"  {msg[1]}: {msg[2]} messages")

# Ask for confirmation
print("\n" + "="*60)
print("RECOMMENDATION:")
print("- Duplicates in broadcast_messages might be intentional")
print("- Same phone might receive multiple campaigns/sequences")
print("- Only delete if you're sure these are unintended duplicates")
print("="*60)

cur.close()
conn.close()
