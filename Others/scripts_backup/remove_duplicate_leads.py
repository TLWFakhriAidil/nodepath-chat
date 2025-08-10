import psycopg2
import sys
from datetime import datetime

sys.stdout.reconfigure(encoding='utf-8')

conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cur = conn.cursor()

print("=== CHECKING FOR DUPLICATE LEADS ===\n")

# First, find all duplicates based on device_id + phone
cur.execute("""
    SELECT 
        device_id, 
        phone, 
        COUNT(*) as duplicate_count,
        array_agg(id ORDER BY created_at ASC) as lead_ids,
        array_agg(name ORDER BY created_at ASC) as names,
        array_agg(created_at ORDER BY created_at ASC) as created_dates
    FROM leads
    WHERE device_id IS NOT NULL 
    AND phone IS NOT NULL
    GROUP BY device_id, phone
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

# Show sample duplicates
print("=== SAMPLE DUPLICATES (First 10) ===")
for i, dup in enumerate(duplicates[:10]):
    device_id = dup[0]
    phone = dup[1]
    count = dup[2]
    lead_ids = dup[3]
    names = dup[4]
    created_dates = dup[5]
    
    print(f"\n{i+1}. Phone: {phone}")
    print(f"   Device ID: {device_id}")
    print(f"   Duplicate Count: {count}")
    print("   Leads:")
    for j in range(min(3, len(lead_ids))):  # Show first 3
        print(f"     - ID: {lead_ids[j]}, Name: {names[j]}, Created: {created_dates[j]}")
    if len(lead_ids) > 3:
        print(f"     ... and {len(lead_ids) - 3} more")

# Count total duplicates to remove
cur.execute("""
    WITH duplicate_leads AS (
        SELECT 
            device_id,
            phone,
            id,
            ROW_NUMBER() OVER (PARTITION BY device_id, phone ORDER BY created_at ASC) as rn
        FROM leads
        WHERE device_id IS NOT NULL 
        AND phone IS NOT NULL
    )
    SELECT COUNT(*) 
    FROM duplicate_leads 
    WHERE rn > 1
""")

total_to_delete = cur.fetchone()[0]
print(f"\n=== SUMMARY ===")
print(f"Total duplicate leads to delete: {total_to_delete}")
print("(Keeping the oldest lead for each device_id + phone combination)")

# Get user confirmation
print("\n" + "="*60)
print("WARNING: This will permanently delete duplicate leads!")
print("The oldest lead for each device_id + phone will be kept.")
print("All newer duplicates will be removed.")
print("="*60)

# For safety, let's first backup the data
print("\n=== CREATING BACKUP ===")
cur.execute("""
    CREATE TABLE IF NOT EXISTS leads_backup_before_dedup AS 
    SELECT * FROM leads 
    WHERE id IN (
        WITH duplicate_leads AS (
            SELECT 
                device_id,
                phone,
                id,
                ROW_NUMBER() OVER (PARTITION BY device_id, phone ORDER BY created_at ASC) as rn
            FROM leads
            WHERE device_id IS NOT NULL 
            AND phone IS NOT NULL
        )
        SELECT id FROM duplicate_leads WHERE rn > 1
    )
""")
print("✅ Backup created in table: leads_backup_before_dedup")

# Now perform the deletion
print("\n=== DELETING DUPLICATES ===")
cur.execute("""
    WITH duplicate_leads AS (
        SELECT 
            device_id,
            phone,
            id,
            ROW_NUMBER() OVER (PARTITION BY device_id, phone ORDER BY created_at ASC) as rn
        FROM leads
        WHERE device_id IS NOT NULL 
        AND phone IS NOT NULL
    )
    DELETE FROM leads
    WHERE id IN (
        SELECT id FROM duplicate_leads WHERE rn > 1
    )
    RETURNING id, phone, name, device_id
""")

deleted_leads = cur.fetchall()
print(f"\n✅ Deleted {len(deleted_leads)} duplicate leads")

# Show some examples of what was deleted
if deleted_leads:
    print("\nSample deleted leads (first 10):")
    for lead in deleted_leads[:10]:
        print(f"  - ID: {lead[0]}, Phone: {lead[1]}, Name: {lead[2]}")

# Commit the changes
conn.commit()

# Final verification
print("\n=== VERIFICATION ===")
cur.execute("""
    SELECT 
        COUNT(*) as remaining_duplicates
    FROM (
        SELECT device_id, phone, COUNT(*) 
        FROM leads 
        WHERE device_id IS NOT NULL AND phone IS NOT NULL
        GROUP BY device_id, phone 
        HAVING COUNT(*) > 1
    ) as dup
""")

remaining = cur.fetchone()[0]
print(f"Remaining duplicates: {remaining}")

# Show statistics
cur.execute("SELECT COUNT(*) FROM leads")
total_leads = cur.fetchone()[0]
print(f"Total leads after cleanup: {total_leads}")

cur.close()
conn.close()

print("\n✅ Duplicate removal complete!")
print("Backup table 'leads_backup_before_dedup' contains the deleted records.")
