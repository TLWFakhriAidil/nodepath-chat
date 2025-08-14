import psycopg2

conn = psycopg2.connect(
    "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"
)
cur = conn.cursor()

print("=== AGGRESSIVE CLEANUP ===\n")

try:
    # 1. First remove triggers from all leads to stop enrollment
    print("1. Removing triggers from leads to prevent re-enrollment...")
    cur.execute("""
        UPDATE leads 
        SET trigger = NULL 
        WHERE trigger IN ('warm_start', 'cold_start', 'hot_start')
    """)
    leads_updated = cur.rowcount
    print(f"   Updated {leads_updated} leads - removed triggers")
    
    # 2. Delete all sequence contacts with force
    print("\n2. Deleting all sequence_contacts...")
    cur.execute("DELETE FROM sequence_contacts WHERE 1=1")
    sc_deleted = cur.rowcount
    print(f"   Deleted {sc_deleted} sequence_contacts")
    
    # 3. Delete all sequence broadcast messages
    print("\n3. Deleting all sequence broadcast_messages...")
    cur.execute("DELETE FROM broadcast_messages WHERE sequence_id IS NOT NULL")
    bm_deleted = cur.rowcount
    print(f"   Deleted {bm_deleted} broadcast_messages")
    
    # Commit all changes
    conn.commit()
    print("\n4. All changes committed!")
    
    # Verify everything is clean
    print("\n5. Final verification:")
    
    # Check sequence_contacts
    cur.execute("SELECT COUNT(*) FROM sequence_contacts")
    sc_count = cur.fetchone()[0]
    print(f"   sequence_contacts: {sc_count} records")
    
    # Check broadcast_messages
    cur.execute("SELECT COUNT(*) FROM broadcast_messages WHERE sequence_id IS NOT NULL")
    bm_count = cur.fetchone()[0]
    print(f"   broadcast_messages with sequence_id: {bm_count} records")
    
    # Check leads with triggers
    cur.execute("""
        SELECT COUNT(*) 
        FROM leads 
        WHERE trigger IN ('warm_start', 'cold_start', 'hot_start')
    """)
    trigger_count = cur.fetchone()[0]
    print(f"   leads with sequence triggers: {trigger_count} records")
    
    # Show the monitoring view
    print("\n6. Sequence Progress Overview:")
    cur.execute("SELECT * FROM sequence_progress_overview")
    results = cur.fetchall()
    print("   Sequence Name    | Should | Enrolled | Active")
    print("   " + "-" * 45)
    for r in results:
        print(f"   {r[0]:<16} | {r[2]:>6} | {r[3]:>8} | {r[4]:>6}")
    
    print("\n=== CLEANUP COMPLETE ===")
    
    if sc_count == 0 and bm_count == 0 and trigger_count == 0:
        print("\nSUCCESS: All data cleaned and triggers removed!")
        print("The Go application can no longer auto-enroll leads.")
    else:
        print(f"\nWARNING: Still have {sc_count + bm_count} records and {trigger_count} triggers!")
        
except Exception as e:
    conn.rollback()
    print(f"\nERROR: {e}")

finally:
    cur.close()
    conn.close()
