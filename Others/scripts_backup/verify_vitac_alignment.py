import psycopg2
import sys
import io
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')

# Connect to database
conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cursor = conn.cursor()

print("="*100)
print("VERIFYING VITAC TRIGGER → VITAC SEQUENCE ALIGNMENT")
print("="*100)

# 1. Check all VITAC-related triggers
print("\n1️⃣ FINDING ALL VITAC-RELATED TRIGGERS IN LEADS:")
print("-"*80)

cursor.execute("""
    SELECT 
        trigger,
        COUNT(*) as lead_count
    FROM leads
    WHERE trigger LIKE '%VITAC%'
    GROUP BY trigger
    ORDER BY lead_count DESC
""")

vitac_triggers = cursor.fetchall()
print(f"\nVITAC Triggers Found:")
for trigger, count in vitac_triggers:
    print(f"  - {trigger}: {count} leads")

# 2. Check if VITAC trigger leads are in correct sequences
print("\n\n2️⃣ CHECKING SEQUENCE ENROLLMENT FOR VITAC TRIGGERS:")
print("-"*80)

cursor.execute("""
    SELECT 
        l.trigger as lead_trigger,
        s.name as sequence_name,
        CASE 
            WHEN s.name LIKE '%VITAC%' THEN '✅ CORRECT'
            ELSE '❌ WRONG'
        END as alignment,
        COUNT(DISTINCT l.phone) as lead_count,
        COUNT(DISTINCT bm.id) as message_count
    FROM leads l
    JOIN broadcast_messages bm ON bm.recipient_phone = l.phone
    JOIN sequences s ON s.id = bm.sequence_id
    WHERE l.trigger LIKE '%VITAC%'
    GROUP BY l.trigger, s.name
    ORDER BY l.trigger, alignment DESC, lead_count DESC
""")

enrollments = cursor.fetchall()

current_trigger = None
for trigger, seq_name, alignment, lead_count, msg_count in enrollments:
    if trigger != current_trigger:
        print(f"\n\nTrigger: {trigger}")
        print("-"*60)
        current_trigger = trigger
    print(f"  {alignment} → {seq_name}")
    print(f"       Leads: {lead_count}, Messages: {msg_count}")

# 3. Find misaligned cases specifically
print("\n\n3️⃣ MISALIGNED CASES (VITAC triggers in non-VITAC sequences):")
print("-"*80)

cursor.execute("""
    SELECT 
        l.phone,
        l.name,
        l.trigger,
        l.niche,
        s.name as wrong_sequence,
        COUNT(bm.id) as message_count,
        COUNT(CASE WHEN bm.status = 'pending' THEN 1 END) as pending_count
    FROM leads l
    JOIN broadcast_messages bm ON bm.recipient_phone = l.phone
    JOIN sequences s ON s.id = bm.sequence_id
    WHERE l.trigger LIKE '%VITAC%'
    AND s.name NOT LIKE '%VITAC%'
    GROUP BY l.phone, l.name, l.trigger, l.niche, s.name
    ORDER BY message_count DESC
    LIMIT 20
""")

misaligned = cursor.fetchall()
if misaligned:
    print(f"\n{'Phone':<15} {'Name':<20} {'Trigger':<15} {'Wrong Sequence':<30} {'Messages':<10} {'Pending'}")
    print("-"*110)
    for row in misaligned:
        phone, name, trigger, niche, wrong_seq, msg_count, pending = row
        print(f"{phone:<15} {name[:20]:<20} {trigger:<15} {wrong_seq:<30} {msg_count:<10} {pending}")
else:
    print("\n✅ No misaligned cases found!")

# 4. Summary statistics
print("\n\n4️⃣ SUMMARY STATISTICS:")
print("-"*80)

cursor.execute("""
    SELECT 
        COUNT(DISTINCT l.phone) as total_vitac_trigger_leads,
        COUNT(DISTINCT CASE WHEN s.name LIKE '%VITAC%' THEN l.phone END) as correctly_enrolled,
        COUNT(DISTINCT CASE WHEN s.name NOT LIKE '%VITAC%' THEN l.phone END) as wrongly_enrolled
    FROM leads l
    LEFT JOIN broadcast_messages bm ON bm.recipient_phone = l.phone
    LEFT JOIN sequences s ON s.id = bm.sequence_id
    WHERE l.trigger LIKE '%VITAC%'
    AND bm.sequence_id IS NOT NULL
""")

stats = cursor.fetchone()
total, correct, wrong = stats

print(f"\n📊 VITAC Trigger Leads:")
print(f"   - Total with VITAC triggers: {total}")
print(f"   - Correctly in VITAC sequences: {correct}")
print(f"   - Wrongly in non-VITAC sequences: {wrong}")
if total > 0:
    accuracy = (correct / total) * 100
    print(f"   - Accuracy: {accuracy:.1f}%")

# 5. Check reverse - non-VITAC triggers in VITAC sequences
print("\n\n5️⃣ REVERSE CHECK (Non-VITAC triggers in VITAC sequences):")
print("-"*80)

cursor.execute("""
    SELECT 
        l.trigger,
        COUNT(DISTINCT l.phone) as lead_count
    FROM leads l
    JOIN broadcast_messages bm ON bm.recipient_phone = l.phone
    JOIN sequences s ON s.id = bm.sequence_id
    WHERE (l.trigger NOT LIKE '%VITAC%' OR l.trigger IS NULL)
    AND s.name LIKE '%VITAC%'
    GROUP BY l.trigger
    ORDER BY lead_count DESC
    LIMIT 10
""")

reverse_issues = cursor.fetchall()
if reverse_issues:
    print("\nNon-VITAC triggers found in VITAC sequences:")
    for trigger, count in reverse_issues:
        trigger_display = trigger if trigger else "NULL/Empty"
        print(f"  - {trigger_display}: {count} leads")

# 6. Recommendations
print("\n\n6️⃣ RECOMMENDATIONS:")
print("-"*80)

if wrong > 0:
    print(f"\n⚠️  Found {wrong} VITAC trigger leads in wrong sequences!")
    print("   Recommended actions:")
    print("   1. Delete pending messages for these misaligned leads")
    print("   2. Let the DirectBroadcastProcessor re-enroll them correctly")
    
    # Count pending messages to fix
    cursor.execute("""
        SELECT COUNT(*)
        FROM broadcast_messages bm
        JOIN leads l ON l.phone = bm.recipient_phone
        JOIN sequences s ON s.id = bm.sequence_id
        WHERE l.trigger LIKE '%VITAC%'
        AND s.name NOT LIKE '%VITAC%'
        AND bm.status = 'pending'
    """)
    pending_to_fix = cursor.fetchone()[0]
    print(f"\n   Pending messages to delete: {pending_to_fix}")
else:
    print("\n✅ All VITAC trigger leads are correctly enrolled in VITAC sequences!")

conn.close()
print("\n" + "="*100)
print("Verification complete!")
