import psycopg2
import sys
from datetime import datetime

# Set UTF-8 encoding
sys.stdout.reconfigure(encoding='utf-8')

# Database connection
conn_string = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

# Data from the pasted content
# Format: (name, phone, niche, trigger, platform, device_id, user_id)
updates_data = [
    ('60163088644', 'WARMEXSTART', 'Wablas', '315e4f8e-6868-4808-a3df-f75e9fce331f', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('601136533727', 'WARMEXSTART', 'Wablas', '315e4f8e-6868-4808-a3df-f75e9fce331f', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('601127272738', 'WARMEXSTART', 'Wablas', 'b0c6279e-bdff-4efb-bffe-468026e53451', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('601110553174', 'HOTEXSTART', 'Wablas', '315e4f8e-6868-4808-a3df-f75e9fce331f', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60168397681', 'HOTEXSTART', 'Wablas', '315e4f8e-6868-4808-a3df-f75e9fce331f', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60167955299', 'WARMEXSTART', 'Wablas', '315e4f8e-6868-4808-a3df-f75e9fce331f', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60172896927', 'WARMEXSTART', 'Wablas', '315e4f8e-6868-4808-a3df-f75e9fce331f', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('601111971716', 'HOTASMART', 'Wablas', 'b0c6279e-bdff-4efb-bffe-468026e53451', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60167741407', 'COLDASMART', 'Wablas', '1929fc54-2307-43b8-9367-d9b6e4a480ab', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60195925572', 'WARMASMART', 'Wablas', '1929fc54-2307-43b8-9367-d9b6e4a480ab', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('601111221417', 'WARMASMART', 'Wablas', '32f0514b-c865-436d-a181-9a16d7ba174a', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60108923759', 'COLDEXSTART', 'Wablas', 'c2f2ed53-b95d-47ff-a12c-d254059704a4', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60109204997', 'COLDASMART', 'Wablas', '32f0514b-c865-436d-a181-9a16d7ba174a', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60107926964', 'COLDASMART', 'Wablas', 'b96ef563-2278-4c95-bc6f-b9887cecff7a', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60192929936', 'COLDASMART', 'Wablas', 'b96ef563-2278-4c95-bc6f-b9887cecff7a', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('601107926964', 'COLDASMART', 'Wablas', 'b96ef563-2278-4c95-bc6f-b9887cecff7a', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60168487543', 'HOTASMART', 'Wablas', '32f0514b-c865-436d-a181-9a16d7ba174a', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60102306467', 'HOTEXSTART', 'Wablas', '1929fc54-2307-43b8-9367-d9b6e4a480ab', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('601160951136', 'COLDASMART', 'Wablas', '00cc3797-e1b6-4ef5-923f-05d578ce04eb', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('601123530745', 'WARMEXSTART', 'Wablas', 'b2fdd012-8e14-4568-9ccb-c0d89263a8e5', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60127764216', 'COLDEXSTART', 'Wablas', 'b2fdd012-8e14-4568-9ccb-c0d89263a8e5', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60103881075', 'HOTVITAC', 'Wablas', '254af4c7-07e2-47a3-b74a-a195b9d71b2d', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60168136490', 'HOTEXSTART', 'Wablas', 'b2fdd012-8e14-4568-9ccb-c0d89263a8e5', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60127684869', 'WARMEXSTART', 'Wablas', '254af4c7-07e2-47a3-b74a-a195b9d71b2d', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60139084136', 'COLDEXSTART', 'Wablas', 'b2fdd012-8e14-4568-9ccb-c0d89263a8e5', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60179297885', 'HOTASMART', 'Wablas', '1929fc54-2307-43b8-9367-d9b6e4a480ab', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('601119241561', 'HOTVITAC', 'Wablas', '254af4c7-07e2-47a3-b74a-a195b9d71b2d', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60194434732', 'HOTEXSTART', 'Wablas', 'b2fdd012-8e14-4568-9ccb-c0d89263a8e5', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('601118664128', 'WARMASMART', 'Wablas', '1929fc54-2307-43b8-9367-d9b6e4a480ab', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60142924960', 'COLDASMART', 'Wablas', '1929fc54-2307-43b8-9367-d9b6e4a480ab', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60138669552', 'HOTASMART', 'Wablas', '1929fc54-2307-43b8-9367-d9b6e4a480ab', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60104368238', 'WARMASMART', 'Wablas', '1929fc54-2307-43b8-9367-d9b6e4a480ab', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('447710173736', 'COLDEXSTART', 'Wablas', 'b2fdd012-8e14-4568-9ccb-c0d89263a8e5', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60138424834', 'HOTEXSTART', 'Wablas', '1929fc54-2307-43b8-9367-d9b6e4a480ab', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60182735934', 'COLDVITAC', 'Wablas', '254af4c7-07e2-47a3-b74a-a195b9d71b2d', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('601111902551', 'COLDEXSTART', 'Wablas', '4ded249a-65bc-4a50-b3cd-0ee014e1599f', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60103961217', 'HOTEXSTART', 'Wablas', '1929fc54-2307-43b8-9367-d9b6e4a480ab', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60163284220', 'HOTEXSTART', 'Wablas', '254af4c7-07e2-47a3-b74a-a195b9d71b2d', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60194066388', 'COLDVITAC', 'Wablas', '254af4c7-07e2-47a3-b74a-a195b9d71b2d', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('601169244724', 'COLDVITAC', 'Wablas', '254af4c7-07e2-47a3-b74a-a195b9d71b2d', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60199631232', 'HOTASMART', 'Wablas', '8b28627f-39c5-493b-b85b-a873342ac954', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60193747511', 'HOTASMART', 'Wablas', '1929fc54-2307-43b8-9367-d9b6e4a480ab', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60189667126', 'HOTEXSTART', 'Wablas', '0c8fc847-b62b-44f8-b24e-143d87ef9c32', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60176963867', 'WARMASMART', 'Wablas', 'a9cc6862-5318-4186-8bb5-a4fcb7a62acf', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60176082925', 'WARMASMART', 'Wablas', '1929fc54-2307-43b8-9367-d9b6e4a480ab', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60194300848', 'HOTASMART', 'Wablas', '8b28627f-39c5-493b-b85b-a873342ac954', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60107014722', 'WARMVITAC', 'Wablas', '8badb299-f1d1-493a-bddf-84cbaba1273b', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('601160812161', 'HOTASMART', 'Wablas', '1929fc54-2307-43b8-9367-d9b6e4a480ab', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60189149432', 'HOTASMART', 'Wablas', '1929fc54-2307-43b8-9367-d9b6e4a480ab', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60149207292', 'COLDASMART', 'Wablas', '7bc49503-c733-45c2-98df-801f094964f4', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60196795277', 'HOTASMART', 'Wablas', '1929fc54-2307-43b8-9367-d9b6e4a480ab', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60175686015', 'HOTEXSTART', 'Wablas', '9a100c0b-956e-4f4b-85af-1fb562e49133', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60138478914', 'HOTVITAC', 'Wablas', '1929fc54-2307-43b8-9367-d9b6e4a480ab', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('601170377471', 'COLDVITAC', 'Wablas', '8badb299-f1d1-493a-bddf-84cbaba1273b', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('601128544779', 'WARMASMART', 'Wablas', '1929fc54-2307-43b8-9367-d9b6e4a480ab', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60199190233', 'COLDVITAC', 'Wablas', '102a5012-eaf1-456b-a7cf-2a29746e7048', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60138674679', 'COLDEXSTART', 'Wablas', '0c8fc847-b62b-44f8-b24e-143d87ef9c32', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60163130486', 'HOTASMART', 'Wablas', '1505fd77-7cfc-4e27-9563-207855f62b13', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60104650611', 'HOTEXSTART', 'Wablas', '8b28627f-39c5-493b-b85b-a873342ac954', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60126163475', 'WARMASMART', 'Wablas', '8b28627f-39c5-493b-b85b-a873342ac954', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('601139864382', 'COLDVITAC', 'Wablas', 'cad4dbd9-1a60-430b-aa74-04542b96dc4f', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60194614981', 'COLDVITAC', 'Wablas', 'cad4dbd9-1a60-430b-aa74-04542b96dc4f', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60139625066', 'WARMASMART', 'Wablas', '00cc3797-e1b6-4ef5-923f-05d578ce04eb', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60198434767', 'WARMASMART', 'Wablas', '7e847128-61c2-49e9-852e-7b1826ac2ad6', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60124614586', 'COLDASMART', 'Wablas', '00cc3797-e1b6-4ef5-923f-05d578ce04eb', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60135424906', 'WARMASMART', 'Wablas', 'a9cc6862-5318-4186-8bb5-a4fcb7a62acf', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60122991856', 'HOTEXSTART', 'Wablas', '0c8fc847-b62b-44f8-b24e-143d87ef9c32', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60177771768', 'WARMEXSTART', 'Wablas', 'cad4dbd9-1a60-430b-aa74-04542b96dc4f', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('601110671598', 'WARM', 'Wablas', '66e7790a-870c-4b20-bdf0-1e565cdd8d2b', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60142818860', 'WARMVITAC', 'Wablas', '1a22247b-e741-4771-8cad-f0af2657321d', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60108924904', 'WARMVITAC', 'Wablas', '8badb299-f1d1-493a-bddf-84cbaba1273b', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('601116466306', 'HOTEXSTART', 'Wablas', '9770d5f7-ccdc-4bba-8645-499db1de91ae', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60182374783', 'COLDVITAC', 'Wablas', '1a22247b-e741-4771-8cad-f0af2657321d', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60189178179', 'WARMEXSTART', 'Whacenter', '87d2dfb4-4e96-4b2b-b5dc-01c71eef026a', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60178020993', 'WARMEXSTART', 'Whacenter', 'f013fea3-50dc-4750-be83-5ac9a9dba8b6', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60169397227', 'WARMEXSTART', 'Whacenter', '74c4e788-79cd-401d-ab4a-ce32f4f30f63', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('6738913139', 'WARMEXSTART', 'Whacenter', '8043385a-5cd7-4c97-b022-96c348ac2437', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('601112098481', 'WARMEXSTART', 'Whacenter', '4b49c4b4-b9cd-4eb0-ab58-15c475f85e8a', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60182216744', 'COLDEXSTART', 'Whacenter', 'b6934b25-394a-4919-94e9-56b1c8bb19a6', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60147969572', 'WARMEXSTART', 'Whacenter', '39b06136-0fa3-4feb-8214-8e6accf447f5', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('601116064866', 'COLDASMART', 'Wablas', '1929fc54-2307-43b8-9367-d9b6e4a480ab', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('601161988629', 'WARMEXSTART', 'Whacenter', '39b06136-0fa3-4feb-8214-8e6accf447f5', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60145381612', 'HOTEXSTART', 'Whacenter', 'b6934b25-394a-4919-94e9-56b1c8bb19a6', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('601126152825', 'WARMEXSTART', 'Whacenter', '9b59d373-6d08-411d-acb2-900fa2dc98a4', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60137960804', 'COLDVITAC', 'Wablas', '8badb299-f1d1-493a-bddf-84cbaba1273b', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('601121998288', 'COLDEXSTART', 'Whacenter', '4a435a00-91c2-48ff-84e5-4aaa4be3f801', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('601111913725', 'WARMEXSTART', 'Wablas', '315e4f8e-6868-4808-a3df-f75e9fce331f', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60197813221', 'WARMEXSTART', 'Whacenter', 'ae0096f8-6110-4e00-961b-440cc04ec00a', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('601123457115', 'HOTEXSTART', 'Wablas', 'b2fdd012-8e14-4568-9ccb-c0d89263a8e5', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('601169535201', 'HOTEXSTART', 'Whacenter', 'f577089b-916d-451e-bcfb-8438f0af4540', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60126973191', 'HOTEXSTART', 'Whacenter', 'f577089b-916d-451e-bcfb-8438f0af4540', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('601128370055', 'HOTEXSTART', 'Whacenter', '9b59d373-6d08-411d-acb2-900fa2dc98a4', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60179505679', 'HOTEXSTART', 'Whacenter', '8043385a-5cd7-4c97-b022-96c348ac2437', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60105525404', 'WARMEXSTART', 'Whacenter', '9b59d373-6d08-411d-acb2-900fa2dc98a4', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60122631139', 'WARMEXSTART', 'Whacenter', '4b49c4b4-b9cd-4eb0-ab58-15c475f85e8a', 'de078f16-3266-4ab3-8153-a248b015228f'),
    # Additional entries from bottom of data
    ('60168312520', 'WARMEXSTART', 'Whacenter', '4b49c4b4-b9cd-4eb0-ab58-15c475f85e8a', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('60106621742', 'HOTEXSTART', 'Wablas', '315e4f8e-6868-4808-a3df-f75e9fce331f', 'de078f16-3266-4ab3-8153-a248b015228f'),
    ('601139197496', 'WARMEXSTART', 'Whacenter', '39b06136-0fa3-4feb-8214-8e6accf447f5', 'de078f16-3266-4ab3-8153-a248b015228f'),
]

try:
    print("Connecting to PostgreSQL database...")
    conn = psycopg2.connect(conn_string)
    cursor = conn.cursor()
    print("[SUCCESS] Connected successfully!\n")
    
    # Start transaction
    cursor.execute("BEGIN")
    
    update_count = 0
    
    print("Updating leads with triggers and platforms...")
    print("-" * 80)
    
    for phone, trigger, platform, device_id, user_id in updates_data:
        try:
            # Update the lead based on phone, device_id, and user_id
            cursor.execute("""
                UPDATE leads 
                SET trigger = %s, 
                    platform = %s,
                    updated_at = CURRENT_TIMESTAMP
                WHERE phone = %s 
                AND device_id = %s 
                AND user_id = %s
                AND (trigger IS NULL OR trigger = '' OR platform IS NULL OR platform = '')
            """, (trigger, platform, phone, device_id, user_id))
            
            if cursor.rowcount > 0:
                update_count += cursor.rowcount
                print(f"Updated: {phone} -> Trigger: {trigger}, Platform: {platform}")
        except Exception as e:
            print(f"Error updating {phone}: {str(e)}")
    
    # Commit the transaction
    cursor.execute("COMMIT")
    print(f"\n[SUCCESS] Successfully updated {update_count} leads!")
    
    # Verify the updates
    print("\n[VERIFICATION - Summary of Updates]:")
    print("-" * 80)
    
    # Check trigger distribution
    cursor.execute("""
        SELECT trigger, COUNT(*) as count
        FROM leads
        WHERE trigger IS NOT NULL AND trigger != ''
        GROUP BY trigger
        ORDER BY count DESC
        LIMIT 20
    """)
    triggers = cursor.fetchall()
    print("\nTrigger Distribution:")
    for trigger, count in triggers:
        print(f"{trigger:30} | {count:,} leads")
    
    # Check platform distribution
    print("\n[Platform Distribution]:")
    cursor.execute("""
        SELECT platform, COUNT(*) as count
        FROM leads
        WHERE platform IS NOT NULL
        GROUP BY platform
        ORDER BY count DESC
    """)
    platforms = cursor.fetchall()
    for platform, count in platforms:
        print(f"{platform:20} | {count:,} leads")
    
    # Check remaining NULL triggers
    cursor.execute("""
        SELECT COUNT(*) 
        FROM leads 
        WHERE trigger IS NULL OR trigger = ''
    """)
    null_count = cursor.fetchone()[0]
    print(f"\n[Remaining leads without triggers]: {null_count:,}")
    
    cursor.close()
    conn.close()
    print("\n[SUCCESS] Database connection closed!")
    
except Exception as e:
    print(f"[ERROR] {str(e)}")
    import traceback
    traceback.print_exc()
    if 'conn' in locals():
        conn.rollback()
        conn.close()
