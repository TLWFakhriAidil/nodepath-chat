import psycopg2
import sys
import io
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')

# Connect to database
conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cursor = conn.cursor()

print("="*80)
print("FIXING VITAC SEQUENCE TRIGGER MISMATCH")
print("="*80)

# First, analyze the problem
cursor.execute("""
    SELECT 
        l.phone,
        l.name,
        l.trigger as lead_trigger,
        l.niche,
        s.name as assigned_sequence,
        s.trigger as sequence_trigger,
        COUNT(bm.id) as message_count
    FROM leads l
    JOIN broadcast_messages bm ON bm.recipient_phone = l.phone
    JOIN sequences s ON s.id = bm.sequence_id
    WHERE l.niche = 'VITAC'
    AND s.name NOT LIKE '%VITAC%'
    GROUP BY l.phone, l.name, l.trigger, l.niche, s.name, s.trigger
    ORDER BY message_count DESC
    LIMIT 20
""")

mismatches = cursor.fetchall()

if mismatches:
    print(f"\nFound {len(mismatches)} leads with VITAC niche but wrong sequence:")
    for mm in mismatches[:5]:  # Show first 5
        print(f"\n  Phone: {mm[0]}")
        print(f"  Name: {mm[1]}")
        print(f"  Lead Trigger: {mm[2]}")
        print(f"  Niche: {mm[3]}")
        print(f"  Wrong Sequence: {mm[4]} (Trigger: {mm[5]})")
        print(f"  Messages: {mm[6]}")

# Now let's fix the issue
print("\n\nFIXING THE ISSUE...")
print("-"*60)

# Step 1: Delete wrong messages for VITAC leads that haven't been sent yet
cursor.execute("""
    DELETE FROM broadcast_messages bm
    WHERE bm.status = 'pending'
    AND EXISTS (
        SELECT 1 FROM leads l 
        WHERE l.phone = bm.recipient_phone 
        AND l.niche = 'VITAC'
    )
    AND bm.sequence_id IN (
        SELECT id FROM sequences WHERE name NOT LIKE '%VITAC%'
    )
    RETURNING id
""")
deleted_count = cursor.rowcount
conn.commit()
print(f"\n✓ Deleted {deleted_count} pending messages with wrong sequences")

# Step 2: Get the correct VITAC sequences
cursor.execute("""
    SELECT id, name, trigger, is_active
    FROM sequences
    WHERE name LIKE '%VITAC%'
    ORDER BY name
""")
vitac_sequences = cursor.fetchall()

print("\n\nAvailable VITAC Sequences:")
for seq in vitac_sequences:
    status = "ACTIVE" if seq[3] else "INACTIVE"
    print(f"  - {seq[1]} (ID: {seq[0]}, Trigger: {seq[2]}, {status})")

# Step 3: Update leads to have correct triggers based on their journey stage
print("\n\nUpdating lead triggers...")

# For leads with VITAC niche, set appropriate triggers
cursor.execute("""
    UPDATE leads 
    SET trigger = CASE 
        WHEN status = 'new' OR status IS NULL THEN 'COLDVITAC'
        WHEN status = 'warm' THEN 'WARMVITAC'
        WHEN status = 'hot' THEN 'HOTVITAC'
        ELSE 'COLDVITAC'
    END
    WHERE niche = 'VITAC'
    AND (trigger IS NULL OR trigger NOT LIKE '%VITAC%')
    RETURNING phone, trigger
""")
updated_leads = cursor.fetchall()
conn.commit()
print(f"\n✓ Updated {len(updated_leads)} lead triggers")

# Step 4: Fix sequence_stepid for existing VITAC messages
print("\n\nFixing missing sequence_stepid for VITAC messages...")

cursor.execute("""
    UPDATE broadcast_messages bm
    SET sequence_stepid = ss.id
    FROM sequence_steps ss
    WHERE bm.sequence_id = ss.sequence_id
    AND bm.sequence_stepid IS NULL
    AND ss.is_entry_point = true
    AND bm.sequence_id IN (SELECT id FROM sequences WHERE name LIKE '%VITAC%')
    RETURNING bm.id
""")
fixed_stepids = cursor.rowcount
conn.commit()
print(f"\n✓ Fixed {fixed_stepids} messages with missing sequence_stepid")

# Step 5: Create new messages for VITAC leads that need them
print("\n\nChecking for VITAC leads that need sequence enrollment...")

cursor.execute("""
    SELECT 
        l.phone,
        l.name,
        l.trigger,
        l.device_id,
        l.user_id
    FROM leads l
    WHERE l.niche = 'VITAC'
    AND l.trigger LIKE '%VITAC%'
    AND NOT EXISTS (
        SELECT 1 FROM broadcast_messages bm
        WHERE bm.recipient_phone = l.phone
        AND bm.sequence_id IN (SELECT id FROM sequences WHERE name LIKE '%VITAC%')
    )
    LIMIT 100
""")
needs_enrollment = cursor.fetchall()

if needs_enrollment:
    print(f"\nFound {len(needs_enrollment)} VITAC leads that need enrollment")
    
    # This would be handled by the sequence processor
    print("\nThese leads will be enrolled automatically by the sequence processor")
    print("based on their triggers:")
    for lead in needs_enrollment[:5]:
        print(f"  - {lead[0]} ({lead[1]}) - Trigger: {lead[2]}")

# Final verification
print("\n\n" + "="*80)
print("VERIFICATION")
print("="*80)

cursor.execute("""
    SELECT 
        s.name,
        COUNT(DISTINCT bm.recipient_phone) as lead_count,
        COUNT(bm.id) as message_count,
        COUNT(CASE WHEN bm.sequence_stepid IS NOT NULL THEN 1 END) as with_stepid
    FROM sequences s
    LEFT JOIN broadcast_messages bm ON bm.sequence_id = s.id
    WHERE s.name LIKE '%VITAC%'
    GROUP BY s.name
    ORDER BY s.name
""")
final_stats = cursor.fetchall()

print("\nVITAC Sequence Statistics:")
for stat in final_stats:
    print(f"\n{stat[0]}:")
    print(f"  - Unique Leads: {stat[1]}")
    print(f"  - Total Messages: {stat[2]}")
    print(f"  - With Step ID: {stat[3]}")

conn.close()
print("\n✅ VITAC sequence fix completed!")
