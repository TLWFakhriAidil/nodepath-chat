import psycopg2

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

try:
    # Connect to database
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("Connected to PostgreSQL database...")
    print("=" * 60)
    
    # First, show what duplicates exist
    print("\n1. Finding duplicate leads (same device_id + phone)...")
    cur.execute("""
        SELECT device_id, phone, COUNT(*) as count
        FROM leads
        WHERE device_id IS NOT NULL
        GROUP BY device_id, phone
        HAVING COUNT(*) > 1
        ORDER BY count DESC
        LIMIT 10
    """)
    
    duplicates = cur.fetchall()
    print(f"Found {len(duplicates)} groups with duplicates (showing top 10)")
    for device_id, phone, count in duplicates:
        print(f"  Device: {device_id[:8]}..., Phone: {phone}, Count: {count}")
    
    # Count total duplicates
    cur.execute("""
        SELECT COUNT(*) FROM (
            SELECT device_id, phone, COUNT(*) as count
            FROM leads
            WHERE device_id IS NOT NULL
            GROUP BY device_id, phone
            HAVING COUNT(*) > 1
        ) as dup
    """)
    total_groups = cur.fetchone()[0]
    
    print(f"\nTotal duplicate groups: {total_groups}")
    
    # Delete duplicates - keep the one with latest created_at
    print("\n2. Removing duplicates (keeping latest created_at)...")
    
    cur.execute("""
        DELETE FROM leads a
        USING (
            SELECT device_id, phone, MAX(created_at) as max_created
            FROM leads
            WHERE device_id IS NOT NULL
            GROUP BY device_id, phone
            HAVING COUNT(*) > 1
        ) b
        WHERE a.device_id = b.device_id 
        AND a.phone = b.phone 
        AND a.created_at < b.max_created
        RETURNING a.id
    """)
    
    deleted_ids = cur.fetchall()
    deleted_count = len(deleted_ids)
    
    print(f"Deleted {deleted_count} duplicate leads")
    
    # Commit the changes
    conn.commit()
    
    # Verify no more duplicates
    cur.execute("""
        SELECT COUNT(*) FROM (
            SELECT device_id, phone, COUNT(*) as count
            FROM leads
            WHERE device_id IS NOT NULL
            GROUP BY device_id, phone
            HAVING COUNT(*) > 1
        ) as dup
    """)
    remaining = cur.fetchone()[0]
    
    print(f"\n3. Verification: {remaining} duplicate groups remaining")
    
    # Show final lead count
    cur.execute("SELECT COUNT(*) FROM leads")
    final_count = cur.fetchone()[0]
    print(f"\nFinal lead count: {final_count}")
    
    print("=" * 60)
    print("CLEANUP COMPLETED!")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
    if conn:
        conn.rollback()
        conn.close()
