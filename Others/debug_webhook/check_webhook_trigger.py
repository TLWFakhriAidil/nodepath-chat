import psycopg2
import json
from datetime import datetime

# Database connection
conn_string = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    print("Connecting to PostgreSQL database...")
    conn = psycopg2.connect(conn_string)
    cursor = conn.cursor()
    print("Connected successfully!\n")
    
    # 1. Check leads table structure
    print("1. LEADS TABLE STRUCTURE:")
    print("-" * 80)
    cursor.execute("""
        SELECT column_name, data_type, is_nullable, column_default
        FROM information_schema.columns
        WHERE table_name = 'leads'
        ORDER BY ordinal_position
    """)
    columns = cursor.fetchall()
    for col in columns:
        print(f"{col[0]:20} | {col[1]:30} | Nullable: {col[2]:3} | Default: {col[3]}")
    
    # 2. Check if trigger column exists
    print("\n2. CHECKING TRIGGER COLUMN:")
    print("-" * 80)
    cursor.execute("""
        SELECT COUNT(*) 
        FROM information_schema.columns 
        WHERE table_name = 'leads' AND column_name = 'trigger'
    """)
    trigger_exists = cursor.fetchone()[0]
    print(f"Trigger column exists: {'YES' if trigger_exists else 'NO'}")
    
    # 3. Check recent leads with triggers
    print("\n3. RECENT LEADS WITH TRIGGERS:")
    print("-" * 80)
    cursor.execute("""
        SELECT id, name, phone, niche, trigger, platform, created_at
        FROM leads
        WHERE trigger IS NOT NULL AND trigger != ''
        ORDER BY created_at DESC
        LIMIT 10
    """)
    leads = cursor.fetchall()
    if leads:
        for lead in leads:
            print(f"ID: {lead[0]} | Name: {lead[1]:20} | Phone: {lead[2]:15}")
            print(f"   Niche: {lead[3] or 'None':15} | Trigger: {lead[4]:30} | Platform: {lead[5] or 'None'}")
            print(f"   Created: {lead[6]}")
            print()
    else:
        print("No leads found with triggers")
    
    # 4. Check leads without triggers
    print("\n4. RECENT LEADS WITHOUT TRIGGERS:")
    print("-" * 80)
    cursor.execute("""
        SELECT id, name, phone, niche, trigger, platform, created_at
        FROM leads
        WHERE trigger IS NULL OR trigger = ''
        ORDER BY created_at DESC
        LIMIT 5
    """)
    leads = cursor.fetchall()
    if leads:
        for lead in leads:
            print(f"ID: {lead[0]} | Name: {lead[1]:20} | Phone: {lead[2]:15}")
            print(f"   Niche: {lead[3] or 'None':15} | Trigger: {lead[4] or 'NULL'} | Platform: {lead[5] or 'None'}")
            print(f"   Created: {lead[6]}")
            print()
    else:
        print("All leads have triggers")
    
    # 5. Check unique triggers
    print("\n5. UNIQUE TRIGGERS IN USE:")
    print("-" * 80)
    cursor.execute("""
        SELECT DISTINCT trigger, COUNT(*) as count
        FROM leads
        WHERE trigger IS NOT NULL AND trigger != ''
        GROUP BY trigger
        ORDER BY count DESC
    """)
    triggers = cursor.fetchall()
    if triggers:
        for trigger in triggers:
            print(f"Trigger: {trigger[0]:30} | Count: {trigger[1]}")
    else:
        print("No triggers found")
    
    # 6. Check webhook-created leads
    print("\n6. WEBHOOK-CREATED LEADS (by platform):")
    print("-" * 80)
    cursor.execute("""
        SELECT platform, COUNT(*) as count, 
               COUNT(CASE WHEN trigger IS NOT NULL AND trigger != '' THEN 1 END) as with_trigger
        FROM leads
        WHERE platform IS NOT NULL
        GROUP BY platform
        ORDER BY count DESC
    """)
    platforms = cursor.fetchall()
    if platforms:
        for platform in platforms:
            print(f"Platform: {platform[0]:20} | Total: {platform[1]:5} | With Trigger: {platform[2]:5}")
    else:
        print("No platform data found")
    
    # 7. Check sequences that use triggers
    print("\n7. SEQUENCES WITH TRIGGERS:")
    print("-" * 80)
    cursor.execute("""
        SELECT s.name, s.trigger, s.is_active, 
               COUNT(DISTINCT ss.trigger) as step_triggers
        FROM sequences s
        LEFT JOIN sequence_steps ss ON s.id = ss.sequence_id
        WHERE s.trigger IS NOT NULL OR ss.trigger IS NOT NULL
        GROUP BY s.id, s.name, s.trigger, s.is_active
        ORDER BY s.name
    """)
    sequences = cursor.fetchall()
    if sequences:
        for seq in sequences:
            print(f"Sequence: {seq[0]:30} | Trigger: {seq[1] or 'None':20}")
            print(f"   Active: {seq[2]} | Step Triggers: {seq[3]}")
    else:
        print("No sequences with triggers found")
    
    # 8. Check specific lead by phone or name to see webhook behavior
    print("\n8. SAMPLE WEBHOOK-CREATED LEADS:")
    print("-" * 80)
    cursor.execute("""
        SELECT id, name, phone, niche, trigger, platform, device_id, user_id, created_at
        FROM leads
        WHERE platform IS NOT NULL
        ORDER BY created_at DESC
        LIMIT 5
    """)
    samples = cursor.fetchall()
    if samples:
        for sample in samples:
            print(f"ID: {sample[0]}")
            print(f"   Name: {sample[1]} | Phone: {sample[2]}")
            print(f"   Niche: {sample[3] or 'None'} | Trigger: {sample[4] or 'None'} | Platform: {sample[5]}")
            print(f"   Device: {sample[6]} | User: {sample[7]}")
            print(f"   Created: {sample[8]}")
            print()
    
    cursor.close()
    conn.close()
    print("\nAnalysis complete!")
    
except Exception as e:
    print(f"Error: {str(e)}")
    import traceback
    traceback.print_exc()
    if 'conn' in locals():
        conn.close()
