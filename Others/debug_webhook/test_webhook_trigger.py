import requests
import json
from datetime import datetime

# Test webhook endpoint
webhook_url = "https://go-whatsapp-web-multidevice-production-b40f.up.railway.app/webhook/lead/create"

# Test data with trigger
test_lead = {
    "name": "Test Lead with Trigger",
    "phone": "60123456789",
    "niche": "TEST",
    "target_status": "prospect",
    "trigger": "TESTCOLD",
    "platform": "WebhookTest",
    "user_id": "de078f16-3266-4ab3-8153-a248b015228f",  # Using the user_id from the database
    "device_id": "102a5012-eaf1-456b-a7cf-2a29746e7048",  # Using a device_id from the database
    "webhook_key": "your-webhook-key-here"  # You need to replace this with the actual webhook key
}

print("Testing webhook with trigger...")
print(f"URL: {webhook_url}")
print(f"Data: {json.dumps(test_lead, indent=2)}")

try:
    response = requests.post(webhook_url, json=test_lead)
    print(f"\nResponse Status: {response.status_code}")
    print(f"Response Body: {json.dumps(response.json(), indent=2)}")
    
    if response.status_code == 200:
        print("\n✅ SUCCESS: Lead created with trigger!")
        result = response.json().get('results', {})
        print(f"Lead ID: {result.get('lead_id')}")
        print(f"Trigger saved: {result.get('trigger')}")
    else:
        print("\n❌ ERROR: Failed to create lead")
        
except Exception as e:
    print(f"\n❌ ERROR: {str(e)}")

print("\n" + "="*50)
print("WEBHOOK KEY REQUIRED!")
print("="*50)
print("You need to set the WEBHOOK_LEAD_KEY environment variable in Railway")
print("Then replace 'your-webhook-key-here' in this script with the actual key")
