import psycopg2
import sys
import io
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')

# Connect to database
conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cursor = conn.cursor()

print("="*100)
print("FINDING AND REMOVING MESSAGES FOR LEADS WITHOUT TRIGGERS")
print("="*100)

# First, find all leads without triggers that have messages
print("\n1️⃣ ANALYZING THE ISSUE:")
print("-"*80)

cursor.execute("""
    SELECT 
        COUNT(DISTINCT l.phone) as leads_without_trigger,
        COUNT(DISTINCT bm.id) as total_messages,
        COUNT(DISTINCT CASE WHEN bm.status = 'pending' THEN bm.id END) as pending_messages,
        COUNT(DISTINCT CASE WHEN bm.status = 'sent' THEN bm.id END) as sent_messages,
        COUNT(DISTINCT CASE WHEN bm.status = 'failed' THEN bm.id END) as failed_messages
    FROM leads l
    JOIN broadcast_messages bm ON bm.recipient_phone = l.phone
    WHERE (l.trigger IS NULL OR l.trigger = '')
    AND bm.sequence_id IS NOT NULL
""")

stats = cursor.fetchone()
print(f"\n📊 Statistics:")
print(f"   - Leads without triggers that have sequence messages: {stats[0]}")
print(f"   - Total messages for these leads: {stats[1]}")
print(f"   - Pending: {stats[2]}")
print(f"   - Sent: {stats[3]}")
print(f"   - Failed: {stats[4]}")

# Show sample of affected leads
print("\n\n2️⃣ SAMPLE OF AFFECTED LEADS:")
print("-"*80)

cursor.execute("""
    SELECT 
        l.phone,
        l.name,
        l.niche,
        COUNT(bm.id) as message_count,
        STRING_AGG(DISTINCT s.name, ', ') as sequences_enrolled
    FROM leads l
    JOIN broadcast_messages bm ON bm.recipient_phone = l.phone
    LEFT JOIN sequences s ON s.id = bm.sequence_id
    WHERE (l.trigger IS NULL OR l.trigger = '')
    AND bm.sequence_id IS NOT NULL
    GROUP BY l.phone, l.name, l.niche
    ORDER BY message_count DESC
    LIMIT 10
""")

affected_leads = cursor.fetchall()
print(f"\n{'Phone':<15} {'Name':<20} {'Niche':<10} {'Messages':<10} {'Sequences'}")
print("-"*100)
for lead in affected_leads:
    phone, name, niche, msg_count, sequences = lead
    print(f"{phone:<15} {name[:20]:<20} {niche or 'N/A':<10} {msg_count:<10} {sequences[:50]}")

# Get user confirmation
print("\n\n3️⃣ DELETION PLAN:")
print("-"*80)
print("   ⚠️  This will DELETE all sequence messages for leads without triggers")
print("   ⚠️  Only PENDING messages will be deleted (not sent or failed)")
print(f"   ⚠️  Total PENDING messages to delete: {stats[2]}")

# Proceed with deletion
print("\n\n4️⃣ PERFORMING DELETION:")
print("-"*80)

# Delete only pending messages for leads without triggers
cursor.execute("""
    DELETE FROM broadcast_messages bm
    WHERE bm.status = 'pending'
    AND bm.sequence_id IS NOT NULL
    AND EXISTS (
        SELECT 1 FROM leads l 
        WHERE l.phone = bm.recipient_phone 
        AND (l.trigger IS NULL OR l.trigger = '')
    )
    RETURNING id, recipient_phone
""")

deleted_messages = cursor.fetchall()
deleted_count = len(deleted_messages)

# Commit the deletion
conn.commit()

print(f"\n✅ Successfully deleted {deleted_count} pending messages")

# Show summary of deleted messages by phone
if deleted_messages:
    phone_counts = {}
    for msg_id, phone in deleted_messages:
        phone_counts[phone] = phone_counts.get(phone, 0) + 1
    
    print("\n\n5️⃣ DELETION SUMMARY BY PHONE:")
    print("-"*80)
    print(f"{'Phone':<15} {'Messages Deleted'}")
    print("-"*40)
    for phone, count in sorted(phone_counts.items(), key=lambda x: x[1], reverse=True)[:10]:
        print(f"{phone:<15} {count}")
    
    if len(phone_counts) > 10:
        print(f"... and {len(phone_counts) - 10} more phones")

# Final verification
print("\n\n6️⃣ FINAL VERIFICATION:")
print("-"*80)

cursor.execute("""
    SELECT 
        COUNT(DISTINCT l.phone) as remaining_leads,
        COUNT(DISTINCT bm.id) as remaining_messages
    FROM leads l
    JOIN broadcast_messages bm ON bm.recipient_phone = l.phone
    WHERE (l.trigger IS NULL OR l.trigger = '')
    AND bm.sequence_id IS NOT NULL
    AND bm.status = 'pending'
""")

final_stats = cursor.fetchone()
print(f"   Remaining leads without triggers with pending messages: {final_stats[0]}")
print(f"   Remaining pending messages for these leads: {final_stats[1]}")

conn.close()
print("\n" + "="*100)
print("✅ CLEANUP COMPLETE!")
print("="*100)
