import psycopg2

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

def check_database_schema():
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== CHECKING ACTUAL DATABASE SCHEMA ===\n")
    
    try:
        # 1. Check sequence_contacts table structure
        print("1. SEQUENCE_CONTACTS TABLE COLUMNS:")
        cur.execute("""
            SELECT 
                column_name, 
                data_type, 
                is_nullable,
                column_default
            FROM information_schema.columns 
            WHERE table_name = 'sequence_contacts'
            ORDER BY ordinal_position
        """)
        
        columns = cur.fetchall()
        print(f"{'Column Name':<30} {'Data Type':<20} {'Nullable':<10} {'Default':<30}")
        print("-" * 90)
        for col in columns:
            print(f"{col[0]:<30} {col[1]:<20} {col[2]:<10} {col[3] or 'None':<30}")
        
        # 2. Check broadcast_messages table structure
        print("\n\n2. BROADCAST_MESSAGES TABLE COLUMNS:")
        cur.execute("""
            SELECT 
                column_name, 
                data_type, 
                is_nullable,
                column_default
            FROM information_schema.columns 
            WHERE table_name = 'broadcast_messages'
            ORDER BY ordinal_position
        """)
        
        columns = cur.fetchall()
        print(f"{'Column Name':<30} {'Data Type':<20} {'Nullable':<10} {'Default':<30}")
        print("-" * 90)
        for col in columns:
            print(f"{col[0]:<30} {col[1]:<20} {col[2]:<10} {col[3] or 'None':<30}")
        
        # 3. Check indexes on sequence_contacts
        print("\n\n3. INDEXES ON SEQUENCE_CONTACTS:")
        cur.execute("""
            SELECT 
                indexname,
                indexdef
            FROM pg_indexes 
            WHERE tablename = 'sequence_contacts'
        """)
        
        indexes = cur.fetchall()
        for idx in indexes:
            print(f"{idx[0]}:")
            print(f"  {idx[1]}")
        
        # 4. Check constraints on sequence_contacts
        print("\n\n4. CONSTRAINTS ON SEQUENCE_CONTACTS:")
        cur.execute("""
            SELECT 
                conname,
                contype,
                pg_get_constraintdef(oid)
            FROM pg_constraint 
            WHERE conrelid = 'sequence_contacts'::regclass
        """)
        
        constraints = cur.fetchall()
        for con in constraints:
            constraint_type = {
                'p': 'PRIMARY KEY',
                'f': 'FOREIGN KEY',
                'u': 'UNIQUE',
                'c': 'CHECK'
            }.get(con[1], con[1])
            print(f"{con[0]} ({constraint_type}):")
            print(f"  {con[2]}")
        
        # 5. Show sample data from sequence_contacts
        print("\n\n5. SAMPLE DATA FROM SEQUENCE_CONTACTS:")
        cur.execute("""
            SELECT 
                id,
                sequence_id,
                contact_phone,
                contact_name,
                current_step,
                status,
                current_trigger,
                assigned_device_id,
                sequence_stepid,
                next_trigger_time
            FROM sequence_contacts
            ORDER BY contact_name, current_step
            LIMIT 10
        """)
        
        data = cur.fetchall()
        if data:
            print(f"{'ID':<36} {'Phone':<15} {'Name':<10} {'Step':<6} {'Status':<10} {'Assigned Device':<36}")
            print("-" * 120)
            for row in data:
                print(f"{row[0]:<36} {row[2]:<15} {row[3]:<10} {row[4]:<6} {row[5]:<10} {row[7] or 'NULL':<36}")
        
        # 6. Check if there are any extra columns not in the code
        print("\n\n6. CHECKING FOR UNEXPECTED COLUMNS:")
        expected_columns = [
            'id', 'sequence_id', 'contact_phone', 'contact_name', 
            'current_step', 'status', 'current_trigger', 'next_trigger_time',
            'assigned_device_id', 'processing_device_id', 'processing_started_at',
            'completed_at', 'created_at', 'updated_at', 'sequence_stepid',
            'user_id', 'retry_count', 'last_error'
        ]
        
        cur.execute("""
            SELECT column_name 
            FROM information_schema.columns 
            WHERE table_name = 'sequence_contacts'
        """)
        
        actual_columns = [row[0] for row in cur.fetchall()]
        
        unexpected = set(actual_columns) - set(expected_columns)
        missing = set(expected_columns) - set(actual_columns)
        
        if unexpected:
            print(f"UNEXPECTED COLUMNS FOUND: {unexpected}")
        if missing:
            print(f"MISSING COLUMNS: {missing}")
        if not unexpected and not missing:
            print("All columns match expected schema")
            
    except Exception as e:
        print(f"Error: {e}")
    finally:
        cur.close()
        conn.close()

if __name__ == "__main__":
    check_database_schema()
