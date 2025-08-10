import requests
import psycopg2

# Connect to database
conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cursor = conn.cursor()

# Get a sample Wablas device with errors
cursor.execute("""
SELECT ud.jid, ud.device_name
FROM user_devices ud
WHERE ud.id = '315e4f8e-b8f8-48bb-a7f0-31c79e039cfe'
""")
device = cursor.fetchone()

if device:
    token = device[0]
    device_name = device[1]
    
    print(f"Testing Wablas API for device: {device_name}")
    print(f"Token: {token[:20]}...")
    
    # Test Wablas API
    url = "https://my.wablas.com/api/device/info"
    headers = {
        "Authorization": token,
        "Content-Type": "application/x-www-form-urlencoded"
    }
    
    try:
        print("\nTesting API connection...")
        response = requests.get(url, headers=headers, timeout=10)
        
        print(f"Status Code: {response.status_code}")
        print(f"Response: {response.text[:200]}...")
        
        if response.status_code == 200:
            print("\n✅ API connection successful!")
        else:
            print(f"\n❌ API error: {response.status_code}")
            
    except requests.exceptions.Timeout:
        print("\n❌ API request timed out!")
    except requests.exceptions.ConnectionError as e:
        print(f"\n❌ Connection error: {e}")
    except Exception as e:
        print(f"\n❌ Unexpected error: {e}")

cursor.close()
conn.close()
