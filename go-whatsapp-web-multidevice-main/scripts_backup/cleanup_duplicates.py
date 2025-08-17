import psycopg2
from datetime import datetime

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

def cleanup_duplicate_devices():
    try:
        # Connect to database
        conn = psycopg2.connect(DB_URI)
        cur = conn.cursor()
        
        print("Connected to PostgreSQL database...")
        print("=" * 60)
        
        # First, let's see the duplicate situation
        print("\n1. Finding devices with duplicate names...")
        cur.execute("""
            SELECT user_id, device_name, COUNT(*) as count
            FROM user_devices
            GROUP BY user_id, device_name
            HAVING COUNT(*) > 1
            ORDER BY count DESC
        """)
        
        duplicate_groups = cur.fetchall()
        print(f"Found {len(duplicate_groups)} groups of duplicate devices")
        
        total_devices_to_remove = 0
        total_leads_to_update = 0
        
        # Process each duplicate group
        for user_id, device_name, count in duplicate_groups:
            print(f"\nProcessing: {device_name} (User: {user_id}, Count: {count})")
            
            # Get all devices in this duplicate group ordered by created_at
            cur.execute("""
                SELECT id, jid, created_at, updated_at, platform
                FROM user_devices
                WHERE user_id = %s AND device_name = %s
                ORDER BY created_at DESC
            """, (user_id, device_name))
            
            devices = cur.fetchall()
            
            # The first one (latest created_at) is the keeper
            keeper_id = devices[0][0]
            latest_jid = devices[0][1]
            
            # Find the most recent JID by checking updated_at
            for device in devices:
                if device[3] > devices[0][3]:  # updated_at > keeper's updated_at
                    latest_jid = device[1]
            
            print(f"  Keeper device ID: {keeper_id}")
            print(f"  Latest JID: {latest_jid}")
            
            # Update keeper device with latest JID
            cur.execute("""
                UPDATE user_devices 
                SET jid = %s, updated_at = NOW()
                WHERE id = %s
            """, (latest_jid, keeper_id))
            
            # Collect IDs of devices to remove
            devices_to_remove = [device[0] for device in devices[1:]]  # All except keeper
            total_devices_to_remove += len(devices_to_remove)
            
            if devices_to_remove:
                print(f"  Devices to remove: {len(devices_to_remove)}")
                
                # Update leads to point to keeper device
                cur.execute("""
                    UPDATE leads 
                    SET device_id = %s::uuid, updated_at = NOW()
                    WHERE device_id = ANY(%s::uuid[])
                    RETURNING id
                """, (keeper_id, devices_to_remove))
                
                updated_leads = cur.fetchall()
                leads_updated = len(updated_leads)
                total_leads_to_update += leads_updated
                print(f"  Leads updated: {leads_updated}")
                
                # Delete duplicate devices
                cur.execute("""
                    DELETE FROM user_devices 
                    WHERE id = ANY(%s::uuid[])
                """, (devices_to_remove,))
                
                print(f"  Deleted {len(devices_to_remove)} duplicate devices")
        
        # Commit all changes
        conn.commit()
        
        print("\n" + "=" * 60)
        print("CLEANUP COMPLETED!")
        print(f"Total duplicate devices removed: {total_devices_to_remove}")
        print(f"Total leads updated: {total_leads_to_update}")
        print("=" * 60)
        
        # Show final device count
        cur.execute("SELECT COUNT(*) FROM user_devices")
        final_count = cur.fetchone()[0]
        print(f"\nFinal device count: {final_count}")
        
        cur.close()
        conn.close()
        
    except Exception as e:
        print(f"Error: {e}")
        if conn:
            conn.rollback()
            conn.close()

if __name__ == "__main__":
    cleanup_duplicate_devices()
