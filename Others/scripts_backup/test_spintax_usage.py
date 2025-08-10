import psycopg2

DB_URI = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== Testing Spintax in Messages ===\n")
    
    # Check if any messages contain spintax patterns
    print("1. Checking for spintax patterns in broadcast_messages content:")
    cur.execute("""
        SELECT COUNT(*) 
        FROM broadcast_messages 
        WHERE content LIKE '%{%|%}%'
    """)
    spintax_count = cur.fetchone()[0]
    print(f"   Messages with spintax patterns: {spintax_count}")
    
    # Check for greeting patterns
    print("\n2. Checking for greeting patterns:")
    greetings = ['Hi ', 'Hello ', 'Hai ', 'Selamat pagi', 'Pagi ']
    
    for greeting in greetings:
        cur.execute("""
            SELECT COUNT(*) 
            FROM broadcast_messages 
            WHERE content LIKE %s
        """, (greeting + '%',))
        count = cur.fetchone()[0]
        print(f"   Messages starting with '{greeting}': {count}")
    
    # Sample some actual message content
    print("\n3. Sample message content:")
    cur.execute("""
        SELECT content, recipient_name
        FROM broadcast_messages
        WHERE status = 'sent'
        ORDER BY sent_at DESC
        LIMIT 5
    """)
    
    messages = cur.fetchall()
    if messages:
        for i, (content, name) in enumerate(messages):
            print(f"\n   Message {i+1} to '{name or 'No name'}':")
            print(f"   {content[:100]}{'...' if len(content) > 100 else ''}")
    else:
        print("   No sent messages found")
    
    # Check if recipient names are being used
    print("\n4. Checking name usage:")
    cur.execute("""
        SELECT COUNT(*) 
        FROM broadcast_messages 
        WHERE recipient_name IS NOT NULL 
        AND recipient_name != ''
        AND content LIKE '%' || recipient_name || '%'
    """)
    name_count = cur.fetchone()[0]
    print(f"   Messages containing recipient name: {name_count}")
    
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
