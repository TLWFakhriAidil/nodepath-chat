import psycopg2

DB_URI = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== Dividing SCHQ-S09 Leads into Quarters ===\n")
    
    # Get device info
    cur.execute("""
        SELECT id, device_name, platform, status, user_id
        FROM user_devices
        WHERE device_name = 'SCHQ-S09'
    """)
    
    device = cur.fetchone()
    if not device:
        print("Device SCHQ-S09 not found!")
        conn.close()
        exit()
    
    device_id = device[0]
    print(f"Found device: {device[1]}")
    print(f"  ID: {device_id}")
    print(f"  Platform: {device[2]}")
    print(f"  Status: {device[3]}")
    
    # Get all leads for this device
    cur.execute("""
        SELECT phone, name, trigger, niche
        FROM leads
        WHERE device_id = %s
        ORDER BY phone
    """, (device_id,))
    
    leads = cur.fetchall()
    total = len(leads)
    print(f"\nTotal leads for this device: {total}")
    
    if total == 0:
        print("No leads found for this device!")
        conn.close()
        exit()
    
    # Show current trigger distribution
    cur.execute("""
        SELECT trigger, COUNT(*) 
        FROM leads
        WHERE device_id = %s
        GROUP BY trigger
        ORDER BY COUNT(*) DESC
    """, (device_id,))
    
    print("\nCurrent trigger distribution:")
    for row in cur.fetchall():
        print(f"  {row[0]}: {row[1]} leads")
    
    # Calculate division
    quarter = total // 4
    remainder = total % 4
    
    cold_count = quarter + (1 if remainder > 0 else 0)
    warm_count = quarter + (1 if remainder > 1 else 0) 
    hot_count = quarter + (1 if remainder > 2 else 0)
    keep_count = total - cold_count - warm_count - hot_count
    
    print(f"\nDividing {total} leads into quarters:")
    print(f"  COLDEXSTART: {cold_count} leads (25%)")
    print(f"  WARMEXSTART: {warm_count} leads (25%)")
    print(f"  HOTEXSTART: {hot_count} leads (25%)")
    print(f"  Keep original: {keep_count} leads (25%)")
    
    # Update triggers
    print("\nUpdating triggers...")
    
    # Get lead phones as list for easier indexing
    lead_phones = [lead[0] for lead in leads]
    
    # Update first quarter to COLDEXSTART
    for i in range(cold_count):
        cur.execute("""
            UPDATE leads
            SET trigger = 'COLDEXSTART'
            WHERE phone = %s AND device_id = %s
        """, (lead_phones[i], device_id))
    
    # Update second quarter to WARMEXSTART
    for i in range(cold_count, cold_count + warm_count):
        cur.execute("""
            UPDATE leads
            SET trigger = 'WARMEXSTART'
            WHERE phone = %s AND device_id = %s
        """, (lead_phones[i], device_id))
    
    # Update third quarter to HOTEXSTART
    for i in range(cold_count + warm_count, cold_count + warm_count + hot_count):
        cur.execute("""
            UPDATE leads
            SET trigger = 'HOTEXSTART'
            WHERE phone = %s AND device_id = %s
        """, (lead_phones[i], device_id))
    
    conn.commit()
    print("Updates committed successfully!")
    
    # Verify the update
    cur.execute("""
        SELECT trigger, COUNT(*) 
        FROM leads
        WHERE device_id = %s
        GROUP BY trigger
        ORDER BY 
            CASE 
                WHEN trigger = 'COLDEXSTART' THEN 1
                WHEN trigger = 'WARMEXSTART' THEN 2
                WHEN trigger = 'HOTEXSTART' THEN 3
                ELSE 4
            END
    """, (device_id,))
    
    print("\n=== Final trigger distribution ===")
    for row in cur.fetchall():
        print(f"  {row[0]}: {row[1]} leads")
    
    # Show sample leads for each trigger
    print("\n=== Sample leads for each trigger ===")
    
    for trigger in ['COLDEXSTART', 'WARMEXSTART', 'HOTEXSTART']:
        cur.execute("""
            SELECT phone, name
            FROM leads
            WHERE device_id = %s AND trigger = %s
            ORDER BY phone
            LIMIT 3
        """, (device_id, trigger))
        
        results = cur.fetchall()
        if results:
            print(f"\n{trigger} samples:")
            for lead in results:
                print(f"  - {lead[0]} ({lead[1] or 'No name'})")
    
    # Summary
    print("\n=== Summary ===")
    print(f"Device SCHQ-S09 now has leads divided into:")
    print(f"- COLDEXSTART: {cold_count} leads (will enter COLD sequence)")
    print(f"- WARMEXSTART: {warm_count} leads (will enter WARM sequence)")  
    print(f"- HOTEXSTART: {hot_count} leads (will enter HOT sequence)")
    print(f"- Original triggers kept: {keep_count} leads")
    
    conn.close()
    print("\nDone! The leads are now ready for sequence enrollment.")
    
except Exception as e:
    print(f"Error: {e}")
