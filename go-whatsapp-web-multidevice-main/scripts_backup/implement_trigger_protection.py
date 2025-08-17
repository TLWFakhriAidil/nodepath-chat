import psycopg2
import sys
import io
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')

# Connect to database
conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cursor = conn.cursor()

print("="*100)
print("IMPLEMENTING DATABASE PROTECTION AGAINST NULL TRIGGER ENROLLMENTS")
print("="*100)

# First, let's check if the trigger function already exists
cursor.execute("""
    SELECT EXISTS (
        SELECT 1 FROM pg_proc 
        WHERE proname = 'check_lead_trigger_before_enrollment'
    )
""")
function_exists = cursor.fetchone()[0]

if function_exists:
    print("\n✅ Trigger function already exists")
else:
    print("\n📝 Creating trigger function...")
    
    try:
        cursor.execute("""
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
                    -- Log the attempt
                    RAISE NOTICE 'Attempted to enroll lead % without trigger in sequence', NEW.recipient_phone;
                    -- For now, just log - don't block (to avoid breaking existing functionality)
                    -- Later change to: RAISE EXCEPTION 'Cannot enroll lead % in sequence - no trigger set', NEW.recipient_phone;
                END IF;
            END IF;
            RETURN NEW;
        END;
        $$ LANGUAGE plpgsql;
        """)
        conn.commit()
        print("✅ Function created successfully")
    except Exception as e:
        print(f"❌ Error creating function: {e}")
        conn.rollback()

# Check if trigger exists
cursor.execute("""
    SELECT EXISTS (
        SELECT 1 FROM pg_trigger 
        WHERE tgname = 'enforce_trigger_requirement'
    )
""")
trigger_exists = cursor.fetchone()[0]

if trigger_exists:
    print("\n✅ Database trigger already exists")
else:
    print("\n📝 Creating database trigger...")
    
    try:
        cursor.execute("""
        CREATE TRIGGER enforce_trigger_requirement
            BEFORE INSERT ON broadcast_messages
            FOR EACH ROW
            EXECUTE FUNCTION check_lead_trigger_before_enrollment();
        """)
        conn.commit()
        print("✅ Trigger created successfully")
    except Exception as e:
        print(f"❌ Error creating trigger: {e}")
        conn.rollback()

# Now let's update the leads that need triggers
print("\n\n📝 UPDATING LEADS WITHOUT TRIGGERS:")
print("-"*80)

# Update EXSTART leads
cursor.execute("""
    UPDATE leads 
    SET trigger = 'COLDEXSTART'
    WHERE niche = 'EXSTART'
    AND (trigger IS NULL OR trigger = '')
    RETURNING phone
""")
exstart_updated = cursor.rowcount
conn.commit()
print(f"✅ Updated {exstart_updated} EXSTART leads with COLDEXSTART trigger")

# Update VITAC leads
cursor.execute("""
    UPDATE leads 
    SET trigger = 'COLDVITAC'
    WHERE niche = 'VITAC'
    AND (trigger IS NULL OR trigger = '')
    RETURNING phone
""")
vitac_updated = cursor.rowcount
conn.commit()
print(f"✅ Updated {vitac_updated} VITAC leads with COLDVITAC trigger")

# Update ASMART leads
cursor.execute("""
    UPDATE leads 
    SET trigger = 'COLDASMART'
    WHERE niche = 'ASMART'
    AND (trigger IS NULL OR trigger = '')
    RETURNING phone
""")
asmart_updated = cursor.rowcount
conn.commit()
print(f"✅ Updated {asmart_updated} ASMART leads with COLDASMART trigger")

# Check remaining leads without triggers
cursor.execute("""
    SELECT 
        niche,
        COUNT(*) as count
    FROM leads
    WHERE (trigger IS NULL OR trigger = '')
    AND niche IS NOT NULL
    GROUP BY niche
    ORDER BY count DESC
""")

remaining = cursor.fetchall()
if remaining:
    print("\n\n⚠️  REMAINING LEADS WITHOUT TRIGGERS:")
    print("-"*80)
    for niche, count in remaining:
        print(f"  - {niche}: {count} leads")

# Final check - leads with no niche and no trigger
cursor.execute("""
    SELECT COUNT(*)
    FROM leads
    WHERE (trigger IS NULL OR trigger = '')
    AND (niche IS NULL OR niche = '')
""")
no_niche_count = cursor.fetchone()[0]
if no_niche_count > 0:
    print(f"\n⚠️  {no_niche_count} leads have neither niche nor trigger - these need manual review")

print("\n\n✅ PROTECTION MEASURES IMPLEMENTED:")
print("-"*80)
print("1. Database function created to check triggers before enrollment")
print("2. Database trigger created on broadcast_messages table")
print("3. Updated leads with appropriate triggers based on niche")
print("\n📌 The system will now log attempts to enroll leads without triggers")
print("📌 To make it strict, update the function to RAISE EXCEPTION instead of NOTICE")

conn.close()
print("\n" + "="*100)
