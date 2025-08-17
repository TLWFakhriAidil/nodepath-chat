import socket
import psycopg2
import sys

print("Testing Railway PostgreSQL connection...")
print("-" * 50)

# Test 1: DNS Resolution
try:
    ip = socket.gethostbyname('yamanote.proxy.rlwy.net')
    print(f"[OK] DNS Resolution successful: {ip}")
except Exception as e:
    print(f"[FAIL] DNS Resolution failed: {e}")
    sys.exit(1)

# Test 2: Port connectivity
try:
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.settimeout(10)
    result = sock.connect_ex(('yamanote.proxy.rlwy.net', 49914))
    sock.close()
    
    if result == 0:
        print("[OK] Port 49914 is reachable")
    else:
        print(f"[FAIL] Port 49914 is not reachable (error code: {result})")
        print("\nPossible reasons:")
        print("1. Railway database might only accept connections from whitelisted IPs")
        print("2. Your network might be blocking outbound connections to this port")
        print("3. The database might be temporarily unavailable")
except Exception as e:
    print(f"[FAIL] Port test failed: {e}")

# Test 3: PostgreSQL connection
try:
    print("\nAttempting PostgreSQL connection...")
    conn = psycopg2.connect(
        "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway",
        connect_timeout=10
    )
    print("[OK] PostgreSQL connection successful!")
    
    cursor = conn.cursor()
    cursor.execute("SELECT version();")
    version = cursor.fetchone()
    print(f"\nDatabase version: {version[0]}")
    
    cursor.close()
    conn.close()
    
except psycopg2.OperationalError as e:
    print(f"[FAIL] PostgreSQL connection failed: {e}")
    print("\nThis usually means:")
    print("1. The database only accepts connections from Railway's network")
    print("2. You need to connect through Railway CLI or from a deployed app")
    print("3. The connection string might have changed")
    print("4. Your IP might not be whitelisted")
except Exception as e:
    print(f"[FAIL] Unexpected error: {e}")

print("\n" + "=" * 50)
print("Connection test complete.")
