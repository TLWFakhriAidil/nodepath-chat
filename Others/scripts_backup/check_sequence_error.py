import psycopg2
import uuid

DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

# These are the contact IDs from the error
contact_ids = [
    'c2c4f64b-8797-4aa0-abfa-89303e9d4388',
    '671e5061-4985-4081-950d-dcb42887216d'
]

try:
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== CHECKING SEQUENCE CONTACTS ===")
    for contact_id in contact_ids:
        cur.execute("""
            SELECT id, status, current_step, next_trigger_time, sequence_id
            FROM sequence_contacts 
            WHERE id = %s
        """, (contact_id,))
        
        result = cur.fetchone()
        if result:
            print(f"\nContact {contact_id}:")
            print(f"  Status: {result[1]}")
            print(f"  Current Step: {result[2]}")
            print(f"  Next Trigger: {result[3]}")
            print(f"  Sequence ID: {result[4]}")
        else:
            print(f"\nContact {contact_id} not found")
    
    # Check if there are any pending contacts
    print("\n=== PENDING CONTACTS ===")
    cur.execute("""
        SELECT COUNT(*) FROM sequence_contacts WHERE status = 'pending'
    """)
    count = cur.fetchone()[0]
    print(f"Total pending contacts: {count}")
    
    # Check for any database functions that might use updated_at
    print("\n=== CHECKING FOR FUNCTIONS WITH UPDATED_AT ===")
    cur.execute("""
        SELECT proname, prosrc 
        FROM pg_proc 
        WHERE prosrc LIKE '%sequence_contacts%' 
        AND prosrc LIKE '%updated_at%'
    """)
    
    functions = cur.fetchall()
    if functions:
        for func in functions:
            print(f"\nFunction {func[0]} contains updated_at reference!")
            print(f"First 200 chars: {func[1][:200]}...")
    else:
        print("No functions found with updated_at reference")
    
    # Check for any triggers
    print("\n=== CHECKING TRIGGERS ===")
    cur.execute("""
        SELECT tgname, pg_get_triggerdef(oid) 
        FROM pg_trigger 
        WHERE tgrelid = 'sequence_contacts'::regclass
    """)
    
    triggers = cur.fetchall()
    if triggers:
        for trigger in triggers:
            print(f"\nTrigger: {trigger[0]}")
            print(f"Definition: {trigger[1]}")
    else:
        print("No triggers found")
        
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
