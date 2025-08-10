import psycopg2
import sys
import io
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')

# Connect to database
conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cursor = conn.cursor()

print("="*100)
print("ANALYZING SEQUENCE ENROLLMENT LOGIC & PREVENTING NULL TRIGGER ENROLLMENTS")
print("="*100)

# 1. Check current sequence triggers and entry points
print("\n1️⃣ CURRENT SEQUENCE ENTRY POINTS:")
print("-"*80)

cursor.execute("""
    SELECT 
        s.name as sequence_name,
        s.is_active,
        ss.trigger as entry_trigger,
        ss.is_entry_point,
        ss.day_number
    FROM sequences s
    JOIN sequence_steps ss ON ss.sequence_id = s.id
    WHERE ss.is_entry_point = true
    ORDER BY s.name
""")

entry_points = cursor.fetchall()
print(f"\n{'Sequence':<30} {'Active':<10} {'Entry Trigger':<20}")
print("-"*60)
for seq_name, active, trigger, is_entry, day in entry_points:
    status = "YES" if active else "NO"
    print(f"{seq_name:<30} {status:<10} {trigger:<20}")

# 2. Check leads without triggers
print("\n\n2️⃣ LEADS WITHOUT TRIGGERS:")
print("-"*80)

cursor.execute("""
    SELECT 
        COUNT(*) as total_leads,
        COUNT(CASE WHEN trigger IS NULL THEN 1 END) as null_triggers,
        COUNT(CASE WHEN trigger = '' THEN 1 END) as empty_triggers,
        COUNT(CASE WHEN trigger IS NOT NULL AND trigger != '' THEN 1 END) as valid_triggers
    FROM leads
""")

stats = cursor.fetchone()
print(f"\nLead Statistics:")
print(f"  - Total leads: {stats[0]:,}")
print(f"  - NULL triggers: {stats[1]:,}")
print(f"  - Empty triggers: {stats[2]:,}")
print(f"  - Valid triggers: {stats[3]:,}")

# 3. Check if any NULL/empty trigger leads are getting new messages
print("\n\n3️⃣ RECENT ENROLLMENTS FOR NULL/EMPTY TRIGGER LEADS:")
print("-"*80)

cursor.execute("""
    SELECT 
        l.phone,
        l.name,
        l.trigger,
        l.niche,
        COUNT(bm.id) as messages_created,
        MAX(bm.created_at) as last_message_created
    FROM leads l
    JOIN broadcast_messages bm ON bm.recipient_phone = l.phone
    WHERE (l.trigger IS NULL OR l.trigger = '')
    AND bm.sequence_id IS NOT NULL
    AND bm.created_at >= CURRENT_DATE - INTERVAL '7 days'
    GROUP BY l.phone, l.name, l.trigger, l.niche
    ORDER BY last_message_created DESC
    LIMIT 10
""")

recent_enrollments = cursor.fetchall()
if recent_enrollments:
    print("\n⚠️  FOUND RECENT ENROLLMENTS FOR LEADS WITHOUT TRIGGERS!")
    print(f"\n{'Phone':<15} {'Name':<20} {'Niche':<10} {'Messages':<10} {'Last Created'}")
    print("-"*80)
    for phone, name, trigger, niche, count, last_created in recent_enrollments:
        trigger_str = trigger if trigger else "NULL"
        print(f"{phone:<15} {name[:20]:<20} {niche or 'N/A':<10} {count:<10} {last_created}")
else:
    print("\n✅ No recent enrollments found for leads without triggers (last 7 days)")

# 4. Check the source code for enrollment logic
print("\n\n4️⃣ CHECKING ENROLLMENT SOURCES:")
print("-"*80)

# Check if there are any direct inserts bypassing trigger checks
cursor.execute("""
    SELECT 
        DATE(created_at) as creation_date,
        COUNT(DISTINCT recipient_phone) as unique_leads,
        COUNT(*) as messages_created
    FROM broadcast_messages
    WHERE sequence_id IS NOT NULL
    AND recipient_phone IN (
        SELECT phone FROM leads WHERE trigger IS NULL OR trigger = ''
    )
    GROUP BY DATE(created_at)
    ORDER BY creation_date DESC
    LIMIT 7
""")

daily_stats = cursor.fetchall()
if daily_stats:
    print("\nDaily message creation for NULL trigger leads:")
    print(f"\n{'Date':<15} {'Unique Leads':<15} {'Messages'}")
    print("-"*45)
    for date, leads, messages in daily_stats:
        print(f"{str(date):<15} {leads:<15} {messages}")

# 5. Protection recommendations
print("\n\n5️⃣ PROTECTION MECHANISMS NEEDED:")
print("-"*80)

print("\n✅ DATABASE LEVEL PROTECTION:")
print("   1. Add CHECK constraint to prevent NULL trigger enrollments")
print("   2. Create trigger function to validate before insert")

# Create the protection SQL
protection_sql = """
-- 1. Function to check trigger before enrollment
CREATE OR REPLACE FUNCTION check_lead_trigger_before_enrollment()
RETURNS TRIGGER AS $$
BEGIN
    -- Only check for sequence messages
    IF NEW.sequence_id IS NOT NULL THEN
        -- Check if lead has a valid trigger
        IF NOT EXISTS (
            SELECT 1 FROM leads l 
            WHERE l.phone = NEW.recipient_phone 
            AND l.trigger IS NOT NULL 
            AND l.trigger != ''
        ) THEN
            RAISE EXCEPTION 'Cannot enroll lead % in sequence - no trigger set', NEW.recipient_phone;
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 2. Create trigger on broadcast_messages
DROP TRIGGER IF EXISTS enforce_trigger_requirement ON broadcast_messages;
CREATE TRIGGER enforce_trigger_requirement
    BEFORE INSERT ON broadcast_messages
    FOR EACH ROW
    EXECUTE FUNCTION check_lead_trigger_before_enrollment();

-- 3. Add constraint to leads table
ALTER TABLE leads 
ADD CONSTRAINT check_trigger_not_empty 
CHECK (
    trigger IS NOT NULL AND trigger != '' 
    OR status = 'inactive'  -- Allow inactive leads to have no trigger
);
"""

print("\n📝 SQL PROTECTION SCRIPT:")
print("-"*80)
print(protection_sql)

# 6. Check Go code patterns
print("\n\n6️⃣ GO CODE PATTERNS TO CHECK:")
print("-"*80)
print("\n⚠️  Make sure DirectBroadcastProcessor checks for trigger before enrollment:")
print("""
// CORRECT Pattern:
if lead.Trigger != "" && lead.Trigger != nil {
    // Proceed with enrollment
}

// WRONG Pattern:
// Enrolling without checking trigger
""")

# 7. Final summary
print("\n\n7️⃣ SUMMARY & RECOMMENDATIONS:")
print("-"*80)

# Get count of leads that need triggers
cursor.execute("""
    SELECT 
        niche,
        COUNT(*) as count
    FROM leads
    WHERE (trigger IS NULL OR trigger = '')
    AND status != 'inactive'
    GROUP BY niche
    ORDER BY count DESC
    LIMIT 5
""")

needs_triggers = cursor.fetchall()
if needs_triggers:
    print("\n⚠️  Leads that need triggers assigned:")
    for niche, count in needs_triggers:
        niche_str = niche if niche else "No Niche"
        print(f"   - {niche_str}: {count:,} leads")

conn.close()
print("\n" + "="*100)
print("Analysis complete!")
