import psycopg2
import csv
from datetime import datetime
import sys

# Set UTF-8 encoding
sys.stdout.reconfigure(encoding='utf-8')

# Database connection
conn_string = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

# Data from the pasted content - I'll parse the key columns
# Format: lead_id, trigger
trigger_updates = [
    (28574, 'WARMEXSTART'),
    (28570, 'WARMEXSTART'),
    (28563, 'WARMEXSTART'),
    (28539, 'HOTEXSTART'),
    (28523, 'HOTEXSTART'),
    (28522, 'WARMEXSTART'),
    (28507, 'WARMEXSTART'),
    (26516, 'HOTASMART'),
    (26249, 'COLDASMART'),
    (25822, 'WARMASMART'),
    (25424, 'WARMASMART'),
    (25392, 'COLDEXSTART'),
    (25306, 'COLDASMART'),
    (24434, 'COLDASMART'),
    (24425, 'COLDASMART'),
    (24424, 'COLDASMART'),
    (24402, 'HOTASMART'),
    (24212, 'HOTEXSTART'),
    (24211, 'HOTEXSTART'),
    (24110, 'COLDASMART'),
    (23895, 'WARMEXSTART'),
    (23790, 'COLDEXSTART'),
    (23710, 'HOTVITAC'),
    (23677, 'HOTEXSTART'),
    (23343, 'WARMEXSTART'),
    (23289, 'WARMASMART'),
    (23275, 'COLDEXSTART'),
    (23083, 'WARMEXSTART'),
    (22808, 'HOTASMART'),
    (22729, 'HOTVITAC'),
    (22663, 'HOTEXSTART'),
    (22585, 'WARMEXSTART'),
    (22554, 'WARMASMART'),
    (22450, 'COLDASMART'),
    (22400, 'HOTASMART'),
    (22159, 'WARMASMART'),
    (22155, 'COLDEXSTART'),
    (21835, 'HOTEXSTART'),
    (21796, 'COLDASMART'),
    (21762, 'COLDASMART'),
    (21718, 'WARMASMART'),
    (21433, 'COLDVITAC'),
    (21360, 'COLDVITAC'),
    (21333, 'COLDEXSTART'),
    (21190, 'HOTEXSTART'),
    (21184, 'HOTEXSTART'),
    (21111, 'WARMASMART'),
    (21018, 'COLDVITAC'),
    (21017, 'COLDVITAC'),
    (20983, 'COLDVITAC'),
    (20683, 'HOTASMART'),
    (20657, 'HOTASMART'),
    (20640, 'COLDEXSTART'),
    (20424, 'COLDEXSTART'),
    (20421, 'HOTEXSTART'),
    (20374, 'WARMASMART'),
    (20345, 'WARMASMART'),
    (20285, 'HOTASMART'),
    (20216, 'COLDVITAC'),
    (20208, 'WARMVITAC'),
    (20104, 'HOTASMART'),
    (19869, 'WARMASMART'),
    (19850, 'WARMASMART'),
    (19796, 'WARMASMART'),
    (19768, 'HOTASMART'),
    (19717, 'WARMASMART'),
    (18972, 'HOTEXSTART'),
    (18445, 'COLDASMART'),
    (18391, 'HOTASMART'),
    (18278, 'HOTEXSTART'),
    (18188, 'HOTVITAC'),
    (17619, 'COLDVITAC'),
    (17553, 'WARMASMART'),
    (17358, 'WARMASMART'),
    (16899, 'COLDVITAC'),
    (15901, 'COLDVITAC'),
    (15529, 'COLDEXSTART'),
    (15112, 'COLDEXSTART'),
    (15030, 'COLDEXSTART'),
    (14743, 'HOTASMART'),
    (14655, 'HOTEXSTART'),
    (14327, 'COLDASMART'),
    (14170, 'WARMASMART'),
    (14109, 'COLDEXSTART'),
    (13957, 'COLDVITAC'),
    (13956, 'COLDVITAC'),
    (13927, 'WARMASMART'),
    (13838, 'WARMASMART'),
    (13815, 'COLDASMART'),
    (13733, 'COLDEXSTART'),
    (13312, 'WARMASMART'),
    (13177, 'HOTEXSTART'),
    (13114, 'WARMEXSTART'),
    (12771, 'WARM'),
    (12468, 'COLDASMART'),
    (12335, 'WARMASMART'),
    (12141, 'WARMASMART'),
    (12132, 'WARMVITAC'),
    (11958, 'COLDVITAC'),
    (11724, 'HOTVITAC'),
    (11281, 'COLDEXSTART'),
    (11229, 'WARMVITAC'),
    (11171, 'HOTEXSTART'),
    (10976, 'HOTEXSTART'),
    (10975, 'WARMASMART'),
    (10877, 'WARM'),
    (10823, 'COLDVITAC'),
    (10552, 'WARMEXSTART'),
    (10480, 'WARMEXSTART'),
    (10310, 'WARMEXSTART'),
    (10298, 'HOTEXSTART'),
    (10274, 'WARMEXSTART'),
    (10220, 'TIDAK JUMPA'),
    (10090, 'WARMVITAC'),
    (10018, 'WARMVITAC'),
    (9772, 'COLDASMART'),
    (9676, 'WARMEXSTART'),
    (9590, 'WARMEXSTART'),
    (9200, 'WARM'),
    (9174, 'WARMASMART'),
    (8566, 'COLDASMART'),
    (8550, 'HOTEXSTART'),
    (8492, 'WARMEXSTART'),
    (8402, 'WARMEXSTART'),
    (8390, 'WARMEXSTART'),
    (8216, 'WARMEXSTART'),
    (8108, 'COLDEXSTART'),
    (8090, 'COLDEXSTART'),
    (8088, 'COLDEXSTART'),
    (8068, 'COLDEXSTART'),
    (8056, 'COLDEXSTART'),
    (7710, 'WARMEXSTART'),
    (7626, 'COLDASMART'),
    (7610, 'WARMEXSTART'),
    (7418, 'WARMVITAC'),
    (7414, 'HOTEXSTART'),
    (7412, 'WARMEXSTART'),
    (7105, 'HOTASMART'),
    (6625, 'COLDEXSTART'),
    (6255, 'HOTEXSTART'),
    (6022, 'COLDVITAC'),
    (5868, 'WARMVITAC'),
    (5836, 'WARMEXSTART'),
    (5658, 'WARMASMART'),
    (5554, 'COLDVITAC'),
    (2227, 'WARMEXSTART'),
    (884, 'COLDEXSTART'),
    (796, 'COLDVITAC'),
    (736, 'WARMEXSTART'),
    (732, 'WARMEXSTART'),
    (506, 'WARMEXSTART'),
    (490, 'HOTVITAC'),
    (436, 'COLDASMART'),
    (318, 'HOTASMART'),
    (288, 'HOTEXSTART'),
    (212, 'COLDEXSTART'),
    (202, 'HOTEXSTART'),
    (200, 'HOTEXSTART'),
    (182, 'HOTEXSTART'),
    (172, 'WARMEXSTART'),
    (160, 'HOTEXSTART'),
    (154, 'HOTEXSTART'),
    (146, 'HOTEXSTART'),
    (76, 'WARMEXSTART'),
    (62, 'WARMEXSTART')
]

try:
    print("Connecting to PostgreSQL database...")
    conn = psycopg2.connect(conn_string)
    cursor = conn.cursor()
    print("Connected successfully!\n")
    
    # First, let's verify these leads exist and have NULL triggers
    print("Verifying leads that need updating...")
    lead_ids = [str(lead_id) for lead_id, _ in trigger_updates]
    
    cursor.execute(f"""
        SELECT id, name, phone, trigger
        FROM leads
        WHERE id IN ({','.join(lead_ids)})
        AND (trigger IS NULL OR trigger = '')
        ORDER BY id
    """)
    
    leads_to_update = cursor.fetchall()
    print(f"Found {len(leads_to_update)} leads with NULL/empty triggers that need updating\n")
    
    # Create a mapping for quick lookup
    update_map = dict(trigger_updates)
    
    # Update each lead
    update_count = 0
    print("Updating leads...")
    for lead_id, name, phone, current_trigger in leads_to_update:
        if lead_id in update_map:
            new_trigger = update_map[lead_id]
            try:
                cursor.execute("""
                    UPDATE leads
                    SET trigger = %s, updated_at = %s
                    WHERE id = %s
                """, (new_trigger, datetime.now(), lead_id))
                update_count += 1
                # Safely print with encoding handling
                safe_name = name.encode('ascii', 'replace').decode('ascii')
                print(f"Updated Lead ID {lead_id}: {safe_name} ({phone}) -> Trigger: {new_trigger}")
            except Exception as e:
                print(f"Error updating lead {lead_id}: {str(e)}")
    
    # Commit the changes
    conn.commit()
    print(f"\n[SUCCESS] Successfully updated {update_count} leads with triggers!")
    
    # Show summary of updates by trigger type
    print("\n[SUMMARY OF UPDATES BY TRIGGER TYPE]:")
    print("-" * 50)
    cursor.execute("""
        SELECT trigger, COUNT(*) as count
        FROM leads
        WHERE id IN ({})
        GROUP BY trigger
        ORDER BY count DESC
    """.format(','.join(lead_ids)))
    
    summary = cursor.fetchall()
    for trigger, count in summary:
        print(f"{trigger:20} | {count:5} leads")
    
    # Verify the updates
    print("\n[VERIFICATION - Sample Updated Leads]:")
    print("-" * 50)
    cursor.execute("""
        SELECT id, name, phone, trigger
        FROM leads
        WHERE id IN ({})
        AND trigger IS NOT NULL
        ORDER BY id DESC
        LIMIT 10
    """.format(','.join(lead_ids[:20])))  # Check first 20 IDs
    
    samples = cursor.fetchall()
    for lead_id, name, phone, trigger in samples:
        safe_name = name.encode('ascii', 'replace').decode('ascii')
        print(f"ID: {lead_id} | {safe_name:20} | {phone:15} | Trigger: {trigger}")
    
    cursor.close()
    conn.close()
    
    print("\n" + "="*60)
    print("Update complete! Triggers have been restored.")
    print("="*60)
    
except Exception as e:
    print(f"Error: {str(e)}")
    import traceback
    traceback.print_exc()
    if 'conn' in locals():
        conn.rollback()
        conn.close()
