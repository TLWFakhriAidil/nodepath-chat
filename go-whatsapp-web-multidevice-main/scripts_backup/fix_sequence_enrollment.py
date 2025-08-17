import psycopg2
import sys

# Set UTF-8 encoding
sys.stdout.reconfigure(encoding='utf-8')

# Connect to database
conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cursor = conn.cursor()

print("=== FIXING SEQUENCE ENROLLMENT ISSUES ===")

# 1. First, check incorrect enrollments
print("\n1. CHECKING INCORRECT ENROLLMENTS")
cursor.execute("""
    SELECT 
        bm.id,
        bm.recipient_phone,
        l.niche,
        l.trigger,
        s.name as sequence_name,
        bm.status
    FROM broadcast_messages bm
    JOIN leads l ON l.phone = bm.recipient_phone
    JOIN sequences s ON s.id = bm.sequence_id
    WHERE bm.status = 'pending'
    AND (
        (l.niche = 'VITAC' AND s.name NOT LIKE '%VITAC%') OR
        (l.niche = 'EXAMA' AND s.name NOT LIKE '%EXAMA%' AND s.name NOT LIKE '%EXSTART%') OR
        (l.niche = 'ASMART' AND s.name NOT LIKE '%ASMART%')
    )
    LIMIT 20
""")

incorrect = cursor.fetchall()
print(f"Found {len(incorrect)} incorrect enrollments (showing first 20)")
for inc in incorrect:
    print(f"  Phone: {inc[1]}, Niche: {inc[2]}, Enrolled in: {inc[4]}")

# 2. Fix leads with niche but no trigger
print("\n2. SETTING TRIGGERS FOR LEADS WITHOUT TRIGGERS")

# Map niche to appropriate starting triggers
niche_trigger_map = {
    'VITAC': 'COLDVITAC',
    'EXAMA': 'COLDEXSTART', 
    'ASMART': 'COLDASMART'
}

for niche, trigger in niche_trigger_map.items():
    # Count leads needing update
    cursor.execute("""
        SELECT COUNT(*) 
        FROM leads 
        WHERE niche = %s 
        AND (trigger IS NULL OR trigger = '')
    """, (niche,))
    count = cursor.fetchone()[0]
    
    if count > 0:
        print(f"\nUpdating {count} {niche} leads to trigger {trigger}")
        
        # Update only leads that haven't been enrolled yet
        cursor.execute("""
            UPDATE leads 
            SET trigger = %s 
            WHERE niche = %s 
            AND (trigger IS NULL OR trigger = '')
            AND NOT EXISTS (
                SELECT 1 FROM broadcast_messages bm
                WHERE bm.recipient_phone = leads.phone
                AND bm.status IN ('pending', 'sent')
            )
        """, (trigger, niche))
        
        updated = cursor.rowcount
        conn.commit()
        print(f"  Updated {updated} leads")

# 3. Clean up incorrect pending messages
print("\n3. CLEANING UP INCORRECT PENDING MESSAGES")

# Delete incorrect enrollments (optional - comment out if you want to keep them)
delete_incorrect = input("\nDelete incorrect enrollments? (y/n): ").lower() == 'y'

if delete_incorrect:
    cursor.execute("""
        DELETE FROM broadcast_messages 
        WHERE id IN (
            SELECT bm.id
            FROM broadcast_messages bm
            JOIN leads l ON l.phone = bm.recipient_phone
            JOIN sequences s ON s.id = bm.sequence_id
            WHERE bm.status = 'pending'
            AND (
                (l.niche = 'VITAC' AND s.name NOT LIKE '%VITAC%') OR
                (l.niche = 'EXAMA' AND s.name NOT LIKE '%EXAMA%' AND s.name NOT LIKE '%EXSTART%') OR
                (l.niche = 'ASMART' AND s.name NOT LIKE '%ASMART%')
            )
        )
    """)
    
    deleted = cursor.rowcount
    conn.commit()
    print(f"Deleted {deleted} incorrect enrollments")

# 4. Verify the fix
print("\n4. VERIFICATION")

# Check lead 601119667332 specifically
cursor.execute("""
    SELECT name, phone, niche, trigger
    FROM leads
    WHERE phone = '601119667332'
""")
lead = cursor.fetchone()
if lead:
    print(f"\nSpecific lead 601119667332:")
    print(f"  Name: {lead[0]}")
    print(f"  Niche: {lead[2]}")
    print(f"  Trigger: {lead[3]}")

# Show sequence trigger mappings
print("\n5. SEQUENCE TRIGGER MAPPINGS")
cursor.execute("""
    SELECT 
        s.name,
        ss.trigger,
        ss.is_entry_point
    FROM sequences s
    JOIN sequence_steps ss ON s.id = ss.sequence_id
    WHERE ss.is_entry_point = true
    ORDER BY s.name
""")

mappings = cursor.fetchall()
print("\nEntry point triggers:")
for m in mappings:
    print(f"  {m[0]}: {m[1]}")

conn.close()
print("\n✅ Fix complete!")
