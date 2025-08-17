import psycopg2
import sys
from datetime import datetime

sys.stdout.reconfigure(encoding='utf-8')

conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cur = conn.cursor()

print("=== CHECKING FOR DUPLICATES: phone + device_id + sequence_stepid ===\n")

# Find duplicates based on recipient_phone + device_id + sequence_stepid
cur.execute("""
    SELECT 
        recipient_phone,
        device_id,
        sequence_stepid,
        COUNT(*) as duplicate_count,
        array_agg(id ORDER BY created_at ASC) as message_ids,
        array_agg(status ORDER BY created_at ASC) as statuses,
        array_agg(created_at ORDER BY created_at ASC) as created_dates,
        array_agg(scheduled_at ORDER BY created_at ASC) as scheduled_dates,
        array_agg(error_message ORDER BY created_at ASC) as errors
    FROM broadcast_messages
    WHERE device_id IS NOT NULL 
    AND recipient_phone IS NOT NULL
    AND sequence_stepid IS NOT NULL
    GROUP BY recipient_phone, device_id, sequence_stepid
    HAVING COUNT(*) > 1
    ORDER BY COUNT(*) DESC
    LIMIT 100
""")

duplicates = cur.fetchall()
print(f"Found {len(duplicates)} duplicate combinations (same phone+device+sequence_step)\n")

if not duplicates:
    print("✅ No duplicates found! Each sequence step is sent only once per phone+device.")
    cur.close()
    conn.close()
    exit()

# Show sample duplicates
print("=== SAMPLE DUPLICATES (First 20) ===")
print("These are TRUE duplicates - same sequence step sent multiple times to same person\n")

for i, dup in enumerate(duplicates[:20]):
    phone = dup[0]
    device_id = dup[1]
    step_id = dup[2]
    count = dup[3]
    message_ids = dup[4]
    statuses = dup[5]
    created_dates = dup[6]
    scheduled_dates = dup[7]
    errors = dup[8]
    
    print(f"{i+1}. Phone: {phone}")
    print(f"   Device: {device_id[:8]}...")
    print(f"   Sequence Step: {step_id[:8]}...")
    print(f"   Duplicate Count: {count}")
    print("   Messages:")
    
    for j in range(min(count, 3)):
        error_info = f", Error: {errors[j]}" if errors[j] else ""
        print(f"     - ID: {message_ids[j][:8]}..., Status: {statuses[j]}{error_info}")
        print(f"       Created: {created_dates[j]}, Scheduled: {scheduled_dates[j]}")
    
    if count > 3:
        print(f"     ... and {count - 3} more")
    print()

# Get statistics about these duplicates
print("=== DUPLICATE STATISTICS ===")

# By status
cur.execute("""
    WITH duplicate_messages AS (
        SELECT 
            recipient_phone,
            device_id,
            sequence_stepid,
            id,
            status,
            ROW_NUMBER() OVER (
                PARTITION BY recipient_phone, device_id, sequence_stepid 
                ORDER BY created_at ASC
            ) as rn
        FROM broadcast_messages
        WHERE device_id IS NOT NULL 
        AND recipient_phone IS NOT NULL
        AND sequence_stepid IS NOT NULL
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

# Count total to delete
cur.execute("""
    WITH duplicate_messages AS (
        SELECT 
            recipient_phone,
            device_id,
            sequence_stepid,
            id,
            status,
            ROW_NUMBER() OVER (
                PARTITION BY recipient_phone, device_id, sequence_stepid 
                ORDER BY 
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
        AND sequence_stepid IS NOT NULL
    )
    SELECT COUNT(*) 
    FROM duplicate_messages 
    WHERE rn > 1
""")

total_to_delete = cur.fetchone()[0]

print(f"\n=== REMOVAL PLAN ===")
print(f"Total TRUE duplicate messages to remove: {total_to_delete}")
print("Strategy: Keep the BEST status (sent > pending > queued > failed) or oldest if same status")

# Ask for confirmation
response = input("\nDo you want to DELETE these duplicate sequence messages? (yes/no): ")

if response.lower() == 'yes':
    # Create backup first
    print("\n=== CREATING BACKUP ===")
    cur.execute("""
        CREATE TABLE IF NOT EXISTS broadcast_messages_backup_sequence_dups AS 
        SELECT bm.* 
        FROM broadcast_messages bm
        WHERE id IN (
            WITH duplicate_messages AS (
                SELECT 
                    recipient_phone,
                    device_id,
                    sequence_stepid,
                    id,
                    ROW_NUMBER() OVER (
                        PARTITION BY recipient_phone, device_id, sequence_stepid 
                        ORDER BY 
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
                AND sequence_stepid IS NOT NULL
            )
            SELECT id FROM duplicate_messages WHERE rn > 1
        )
    """)
    print("✅ Backup created: broadcast_messages_backup_sequence_dups")
    
    # Delete duplicates
    print("\n=== DELETING DUPLICATES ===")
    cur.execute("""
        WITH duplicate_messages AS (
            SELECT 
                recipient_phone,
                device_id,
                sequence_stepid,
                id,
                ROW_NUMBER() OVER (
                    PARTITION BY recipient_phone, device_id, sequence_stepid 
                    ORDER BY 
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
            AND sequence_stepid IS NOT NULL
        )
        DELETE FROM broadcast_messages
        WHERE id IN (
            SELECT id FROM duplicate_messages WHERE rn > 1
        )
        RETURNING id, recipient_phone, status
    """)
    
    deleted = cur.fetchall()
    print(f"\n✅ Deleted {len(deleted)} duplicate sequence messages")
    
    # Show sample of what was deleted
    if deleted:
        print("\nSample deleted messages (first 10):")
        for msg in deleted[:10]:
            print(f"  - ID: {msg[0][:8]}..., Phone: {msg[1]}, Status: {msg[2]}")
    
    conn.commit()
    
    # Verify
    print("\n=== VERIFICATION ===")
    cur.execute("""
        SELECT COUNT(*) 
        FROM (
            SELECT recipient_phone, device_id, sequence_stepid
            FROM broadcast_messages
            WHERE device_id IS NOT NULL 
            AND recipient_phone IS NOT NULL
            AND sequence_stepid IS NOT NULL
            GROUP BY recipient_phone, device_id, sequence_stepid
            HAVING COUNT(*) > 1
        ) as remaining
    """)
    
    remaining = cur.fetchone()[0]
    print(f"Remaining duplicates: {remaining}")
    
    print("\n✅ Duplicate sequence messages have been cleaned up!")
    print("Backup table 'broadcast_messages_backup_sequence_dups' contains deleted records.")
else:
    print("\nDeletion cancelled. No changes made.")

cur.close()
conn.close()
