# Add encoding to handle Unicode
import sys
import io
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')

import psycopg2
import re

# Connect to database
conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cursor = conn.cursor()

print("="*80)
print("INVESTIGATING WHATSAPP MULTI-DEVICE ISSUES")
print("="*80)

# Issue 1: Check the phone number trigger mismatch
print("\n1. CHECKING PHONE 601119667332 - TRIGGER MISMATCH ISSUE")
print("-"*60)

# Get lead info
cursor.execute("""
    SELECT id, name, phone, trigger, niche, device_id, user_id
    FROM leads 
    WHERE phone = '601119667332'
""")
lead = cursor.fetchone()

if lead:
    lead_id, name, phone, trigger, niche, device_id, user_id = lead
    print(f"Lead found:")
    print(f"  - ID: {lead_id}")
    print(f"  - Name: {name}")
    print(f"  - Phone: {phone}")
    print(f"  - Trigger: {trigger}")
    print(f"  - Niche: {niche}")
    
    # Check messages for this lead
    cursor.execute("""
        SELECT 
            bm.id,
            bm.content,
            bm.sequence_id,
            bm.sequence_stepid,
            bm.status,
            bm.scheduled_at,
            s.name as sequence_name,
            s.trigger as sequence_trigger,
            ss.trigger as step_trigger,
            ss.is_entry_point
        FROM broadcast_messages bm
        LEFT JOIN sequences s ON s.id = bm.sequence_id
        LEFT JOIN sequence_steps ss ON ss.id = bm.sequence_stepid
        WHERE bm.recipient_phone = '601119667332'
        ORDER BY bm.scheduled_at DESC
        LIMIT 10
    """)
    messages = cursor.fetchall()
    
    print(f"\nMessages for this lead ({len(messages)} found):")
    for msg in messages:
        msg_id, content, seq_id, step_id, status, scheduled, seq_name, seq_trigger, step_trigger, is_entry = msg
        print(f"\n  Message: {msg_id}")
        print(f"    - Sequence: {seq_name} (Trigger: {seq_trigger})")
        print(f"    - Step Trigger: {step_trigger} (Entry Point: {is_entry})")
        print(f"    - Status: {status}")
        print(f"    - Scheduled: {scheduled}")
        if content:
            print(f"    - Content preview: {content[:60]}...")

# Check all VITAC sequences and their usage
print("\n\n2. ANALYZING VITAC SEQUENCES VS EXSTART SEQUENCES")
print("-"*60)

cursor.execute("""
    SELECT 
        s.name,
        s.trigger,
        s.is_active,
        COUNT(DISTINCT bm.id) as message_count,
        COUNT(DISTINCT bm.recipient_phone) as lead_count
    FROM sequences s
    LEFT JOIN broadcast_messages bm ON bm.sequence_id = s.id
    WHERE s.name LIKE '%VITAC%' OR s.name LIKE '%Sequence%'
    GROUP BY s.id, s.name, s.trigger, s.is_active
    ORDER BY s.name
""")
sequences = cursor.fetchall()

print("\nSequence Usage Analysis:")
for seq in sequences:
    print(f"\n{seq[0]}:")
    print(f"  - Trigger: {seq[1]}")
    print(f"  - Active: {seq[2]}")
    print(f"  - Messages: {seq[3]}")
    print(f"  - Unique Leads: {seq[4]}")

# Check for wrong sequence assignments
print("\n\n3. CHECKING FOR WRONG SEQUENCE ASSIGNMENTS")
print("-"*60)

cursor.execute("""
    SELECT 
        l.trigger as lead_trigger,
        s.name as sequence_name,
        s.trigger as sequence_trigger,
        COUNT(*) as count
    FROM broadcast_messages bm
    JOIN leads l ON l.phone = bm.recipient_phone
    JOIN sequences s ON s.id = bm.sequence_id
    WHERE l.trigger LIKE '%VITAC%'
    AND s.name NOT LIKE '%VITAC%'
    GROUP BY l.trigger, s.name, s.trigger
""")
mismatches = cursor.fetchall()

if mismatches:
    print("\nFOUND TRIGGER MISMATCHES:")
    for mismatch in mismatches:
        print(f"\n  Lead Trigger: {mismatch[0]}")
        print(f"  Assigned Sequence: {mismatch[1]} (Trigger: {mismatch[2]})")
        print(f"  Affected Count: {mismatch[3]}")
else:
    print("\nNo obvious trigger mismatches found.")

# Check sequence_stepid null values for VITAC
print("\n\n4. CHECKING NULL SEQUENCE_STEPID FOR VITAC MESSAGES")
print("-"*60)

cursor.execute("""
    SELECT 
        s.name,
        COUNT(*) as total_messages,
        COUNT(bm.sequence_stepid) as with_stepid,
        COUNT(*) - COUNT(bm.sequence_stepid) as without_stepid
    FROM broadcast_messages bm
    JOIN sequences s ON s.id = bm.sequence_id
    WHERE s.name LIKE '%VITAC%'
    GROUP BY s.name
""")
stepid_stats = cursor.fetchall()

print("\nSequence Step ID Statistics for VITAC:")
for stat in stepid_stats:
    print(f"\n{stat[0]}:")
    print(f"  - Total Messages: {stat[1]}")
    print(f"  - With Step ID: {stat[2]}")
    print(f"  - Missing Step ID: {stat[3]}")

# Check spintax patterns in sequences vs campaigns
print("\n\n5. ANALYZING SPINTAX USAGE IN SEQUENCES VS CAMPAIGNS")
print("-"*60)

# Check sequence messages for spintax
cursor.execute("""
    SELECT 
        'Sequence' as source,
        COUNT(*) as total,
        COUNT(CASE WHEN content LIKE '%{%|%}%' THEN 1 END) as with_spintax
    FROM sequence_steps
    WHERE content IS NOT NULL
    
    UNION ALL
    
    SELECT 
        'Campaign' as source,
        COUNT(*) as total,
        COUNT(CASE WHEN message LIKE '%{%|%}%' THEN 1 END) as with_spintax
    FROM campaigns
    WHERE message IS NOT NULL
""")
spintax_stats = cursor.fetchall()

print("\nSpintax Usage Statistics:")
for stat in spintax_stats:
    percentage = (stat[2] / stat[1] * 100) if stat[1] > 0 else 0
    print(f"\n{stat[0]}:")
    print(f"  - Total Messages: {stat[1]}")
    print(f"  - With Spintax: {stat[2]} ({percentage:.1f}%)")

# Show sample spintax patterns
cursor.execute("""
    SELECT content 
    FROM sequence_steps 
    WHERE content LIKE '%{%|%}%' 
    LIMIT 3
""")
examples = cursor.fetchall()

if examples:
    print("\nSample Spintax from Sequences:")
    for i, ex in enumerate(examples, 1):
        print(f"\nExample {i}:")
        # Find spintax patterns
        spintax_patterns = re.findall(r'\{[^}]+\}', ex[0])
        for pattern in spintax_patterns[:3]:  # Show first 3 patterns
            print(f"  - {pattern}")

conn.close()
print("\n" + "="*80)
print("Investigation complete!")
