import psycopg2

conn_string = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(conn_string)
    cur = conn.cursor()
    
    print("=== DEVICE OWNERSHIP ISSUE ===")
    
    # Check who owns device d409cadc
    device_id = "d409cadc-75e2-4004-a789-c2bad0b31393"
    cur.execute("""
        SELECT d.id, d.user_id, u.email, d.phone, d.status
        FROM user_devices d
        LEFT JOIN users u ON d.user_id = u.id
        WHERE d.id = %s
    """, (device_id,))
    
    result = cur.fetchone()
    if result:
        print(f"Device {device_id}:")
        print(f"- Owner User ID: {result[1]}")
        print(f"- Owner Email: {result[2]}")
        print(f"- Phone: {result[3]}")
        print(f"- Status: {result[4]}")
    
    # Check campaign 59 owner
    cur.execute("""
        SELECT c.user_id, u.email
        FROM campaigns c
        LEFT JOIN users u ON c.user_id = u.id
        WHERE c.id = 59
    """)
    
    result = cur.fetchone()
    if result:
        print(f"\nCampaign 59:")
        print(f"- Owner User ID: {result[0]}")
        print(f"- Owner Email: {result[1]}")
    
    print("\n=== THE PROBLEM ===")
    print("The campaign is created by one user but the device belongs to another user!")
    print("Campaigns only process messages for devices owned by the campaign creator.")
    
    print("\n=== SOLUTION ===")
    print("Option 1: Create the campaign with the user who owns the device")
    print("Option 2: Transfer the device to the campaign creator")
    print("Option 3: Update the campaign to use the correct user's devices")
    
    # Show which devices the campaign user has
    cur.execute("""
        SELECT id, phone, status, platform
        FROM user_devices
        WHERE user_id = 'de078f16-3266-4ab3-8153-a248b015228f'
    """)
    
    print("\n=== DEVICES OWNED BY CAMPAIGN CREATOR ===")
    for row in cur.fetchall():
        print(f"- {row[0]} | Phone: {row[1]} | Status: {row[2]} | Platform: {row[3]}")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
