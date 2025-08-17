# Simple Lead Creation Webhook

## Endpoint
```
POST /webhook/lead/create
```

## Request Format

### Headers
```
Content-Type: application/json
```

### Request Body (Exactly as you want)
```json
{
  "name": "John Doe",
  "phone": "60123456789",
  "target_status": "prospect",
  "device_id": "device-id",
  "user_id": "user_id",
  "niche": "EXSTART",
  "trigger": "NEWNP"
}
```

### Field Mapping to Database

| JSON Field | Database Column | Required | Description |
|------------|----------------|----------|-------------|
| name | name | Yes | Lead's full name |
| phone | phone | Yes | Phone number |
| target_status | target_status | No | Lead status (prospect/customer) |
| device_id | device_id | Yes | Device ID |
| user_id | user_id | Yes | User ID |
| niche | niche | No | Lead niche/category |
| trigger | trigger | No | Sequence trigger |

## Response

### Success (200)
```json
{
  "status": 200,
  "code": "SUCCESS",
  "message": "Lead created successfully",
  "results": {
    "lead_id": "generated-uuid",
    "name": "John Doe",
    "phone": "60123456789",
    "niche": "EXSTART",
    "trigger": "NEWNP",
    "target_status": "prospect",
    "device_id": "device-id",
    "user_id": "user_id",
    "created_at": "2025-01-15T10:30:00Z"
  }
}
```

### Error - Missing Required Field (400)
```json
{
  "status": 400,
  "code": "VALIDATION_ERROR",
  "message": "Name is required"
}
```

## Example Usage

### cURL
```bash
curl -X POST https://your-app.railway.app/webhook/lead/create \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "phone": "60123456789",
    "target_status": "prospect",
    "device_id": "device-id",
    "user_id": "user_id",
    "niche": "EXSTART",
    "trigger": "NEWNP"
  }'
```

### Node.js
```javascript
const axios = require('axios');

const createLead = async (leadData) => {
  try {
    const response = await axios.post('https://your-app.railway.app/webhook/lead/create', {
      name: leadData.name,
      phone: leadData.phone,
      target_status: "prospect",
      device_id: leadData.deviceId,
      user_id: leadData.userId,
      niche: leadData.niche || "EXSTART",
      trigger: leadData.trigger || "NEWNP"
    });
    
    console.log('Lead created:', response.data);
    return response.data.results.lead_id;
  } catch (error) {
    console.error('Error:', error.response?.data || error.message);
    return null;
  }
};

// Usage
createLead({
  name: "John Doe",
  phone: "60123456789",
  deviceId: "your-device-id",
  userId: "your-user-id",
  niche: "EXSTART",
  trigger: "NEWNP"
});
```

### Python
```python
import requests

def create_lead(lead_data):
    url = "https://your-app.railway.app/webhook/lead/create"
    
    payload = {
        "name": lead_data["name"],
        "phone": lead_data["phone"],
        "target_status": "prospect",
        "device_id": lead_data["device_id"],
        "user_id": lead_data["user_id"],
        "niche": lead_data.get("niche", "EXSTART"),
        "trigger": lead_data.get("trigger", "NEWNP")
    }
    
    try:
        response = requests.post(url, json=payload)
        response.raise_for_status()
        
        result = response.json()
        print(f"Lead created: {result['results']['lead_id']}")
        return result['results']['lead_id']
    except requests.exceptions.RequestException as e:
        print(f"Error: {e}")
        if hasattr(e.response, 'json'):
            print(f"Details: {e.response.json()}")
        return None

# Usage
create_lead({
    "name": "John Doe",
    "phone": "60123456789",
    "device_id": "your-device-id",
    "user_id": "your-user-id",
    "niche": "EXSTART",
    "trigger": "NEWNP"
})
```

## Notes

1. **No authentication required** - The webhook is public
2. **Required fields**: name, phone, device_id, user_id
3. **Optional fields**: target_status, niche, trigger
4. **Only saves what you send** - No default values are added

## Testing Locally

1. Run the app:
```bash
cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
whatsapp.exe rest
```

2. Test with local URL:
```bash
curl -X POST http://localhost:3000/webhook/lead/create \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Lead",
    "phone": "60123456789",
    "target_status": "prospect",
    "device_id": "test-device-id",
    "user_id": "test-user-id",
    "niche": "EXSTART",
    "trigger": "NEWNP"
  }'
```

## Database Table Structure

The webhook creates records in the `leads` table with these columns:
- `id` - Auto-generated UUID
- `name` - From request
- `phone` - From request
- `niche` - From request
- `trigger` - From request
- `target_status` - From request
- `device_id` - From request
- `user_id` - From request
- `created_at` - Current timestamp
- `updated_at` - Current timestamp

All other columns (source, status, email, notes) will be NULL unless explicitly provided in the request.
