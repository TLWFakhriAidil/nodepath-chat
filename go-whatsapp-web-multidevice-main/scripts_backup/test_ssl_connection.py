import psycopg2
import ssl
import time

# Try multiple connection approaches
connection_strings = [
    # With SSL
    "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require",
    # Without SSL
    "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway",
    # With prefer SSL
    "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=prefer"
]

for i, conn_str in enumerate(connection_strings, 1):
    print(f"\nAttempt {i}: Trying connection...")
    print(f"SSL mode: {conn_str.split('sslmode=')[1] if 'sslmode' in conn_str else 'default'}")
    
    try:
        start_time = time.time()
        conn = psycopg2.connect(conn_str, connect_timeout=20)
        elapsed = time.time() - start_time
        
        print(f"[SUCCESS] Connected in {elapsed:.2f} seconds!")
        
        cursor = conn.cursor()
        cursor.execute("SELECT version();")
        version = cursor.fetchone()
        print(f"Database: {version[0][:50]}...")
        
        # Quick table count
        cursor.execute("""
            SELECT COUNT(*) 
            FROM information_schema.tables 
            WHERE table_schema = 'public'
        """)
        table_count = cursor.fetchone()[0]
        print(f"Tables found: {table_count}")
        
        cursor.close()
        conn.close()
        print("\nConnection successful! Use this connection string:")
        print(conn_str)
        break
        
    except psycopg2.OperationalError as e:
        print(f"[FAILED] {str(e)[:100]}...")
    except Exception as e:
        print(f"[ERROR] {type(e).__name__}: {str(e)[:100]}...")

print("\n" + "="*60)
print("Note: Railway databases may restrict connections to:")
print("1. Applications deployed on Railway")
print("2. Railway CLI connections")
print("3. Whitelisted IP addresses")
print("\nIf all connections fail, you may need to:")
print("- Use Railway CLI: railway run psql")
print("- Deploy your app to Railway")
print("- Add your IP to whitelist (if available)")
