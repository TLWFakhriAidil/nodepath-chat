# Lead Creation Webhook Documentation

## Overview
This webhook endpoint allows external services (like your custom WhatsApp bot) to automatically create leads in the WhatsApp Multi-Device system.

## Endpoint
```
POST /webhook/lead/create
```

## Authentication
The webhook uses a simple shared key authentication. Set the key in your environment:

```env
WEBHOOK_LEAD_KEY=your-secret-webhook-key-here
```

## Request Format

### Headers
```
Content-Type: application/json
```

### Request Body
```json
{
  "name": "John Doe",
  "phone": "60123456789",
  "email": "john@example.com",
  "niche": "fitness",
  "source": "whatsapp_bot",
  "target_status": "prospect",
  "trigger": "fitness_start",
  "notes": "Interested in weight loss program",
  "user_id": "your-user-uuid-here",
  "device_id": "optional-device-uuid",
  "webhook_key": "your-secret-webhook-key-here"
}
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | Yes | Lead's full name |
| phone | string | Yes | Phone number (without + or spaces) |
| email | string | No | Email address |
| niche | string | No | Lead category (e.g., fitness, crypto, etc.) |
| source | string | No | Lead source (defaults to "webhook") |
| target_status | string | Yes | Must be either "prospect" or "customer" |
| trigger | string | No | Comma-separated sequence triggers |
| notes | string | No | Additional notes about the lead |
| user_id | string | Yes | UUID of the user who owns this lead |
| device_id | string | No | UUID of the device (must belong to user) |
| webhook_key | string | Yes | Your webhook authentication key |

## Response Format

### Success Response (200)
```json
{
  "status": 200,
  "code": "SUCCESS",
  "message": "Lead created successfully",
  "results": {
    "lead_id": "generated-uuid",
    "name": "John Doe",
    "phone": "60123456789",
    "target_status": "prospect",
    "created_at": "2025-01-15T10:30:00Z"
  }
}
```

### Error Responses

#### Invalid Request Body (400)
```json
{
  "status": 400,
  "code": "BAD_REQUEST",
  "message": "Invalid request body",
  "results": {
    "error": "invalid character 'x' looking for beginning of value"
  }
}
```

#### Validation Error (400)
```json
{
  "status": 400,
  "code": "VALIDATION_ERROR",
  "message": "Validation failed",
  "results": {
    "errors": {
      "name": "required",
      "target_status": "must be one of: prospect customer"
    }
  }
}
```

#### Unauthorized (401)
```json
{
  "status": 401,
  "code": "UNAUTHORIZED",
  "message": "Invalid webhook key"
}
```

#### User Not Found (404)
```json
{
  "status": 404,
  "code": "USER_NOT_FOUND",
  "message": "User not found"
}
```

#### Device Not Found (404)
```json
{
  "status": 404,
  "code": "DEVICE_NOT_FOUND",
  "message": "Device not found or doesn't belong to user"
}
```

#### Server Error (500)
```json
{
  "status": 500,
  "code": "CREATE_FAILED",
  "message": "Failed to create lead",
  "results": {
    "error": "database error details"
  }
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
    "user_id": "your-user-uuid",
    "webhook_key": "your-secret-webhook-key-here"
  }'
```

### Node.js (Axios)
```javascript
const axios = require('axios');

const createLead = async () => {
  try {
    const response = await axios.post('https://your-app.railway.app/webhook/lead/create', {
      name: "John Doe",
      phone: "60123456789",
      email: "john@example.com",
      niche: "fitness",
      source: "whatsapp_bot",
      target_status: "prospect",
      trigger: "fitness_start",
      notes: "Interested in weight loss",
      user_id: "your-user-uuid",
      webhook_key: "your-secret-webhook-key-here"
    });
    
    console.log('Lead created:', response.data);
  } catch (error) {
    console.error('Error:', error.response?.data || error.message);
  }
};
```

### Python
```python
import requests

url = "https://your-app.railway.app/webhook/lead/create"
payload = {
    "name": "John Doe",
    "phone": "60123456789",
    "email": "john@example.com",
    "niche": "fitness",
    "source": "whatsapp_bot",
    "target_status": "prospect",
    "trigger": "fitness_start",
    "notes": "Interested in weight loss",
    "user_id": "your-user-uuid",
    "webhook_key": "your-secret-webhook-key-here"
}

response = requests.post(url, json=payload)
print(response.json())
```

## Integration with Your WhatsApp Bot

When your WhatsApp bot receives a message from a new contact:

1. Extract the contact information
2. Determine the target_status based on your business logic
3. Set appropriate triggers for sequence automation
4. Send the webhook request to create the lead

Example bot integration flow:
```
WhatsApp Message Received → Bot Processes → Webhook Called → Lead Created → Sequences Triggered
```

## Security Considerations

1. **Keep your webhook key secret** - Don't commit it to version control
2. **Use HTTPS** - Railway provides SSL by default
3. **Validate phone numbers** - Ensure they're in the correct format
4. **Rate limiting** - Consider implementing rate limits if needed
5. **IP whitelisting** - Can be added for additional security

## Testing the Webhook

1. First, get your user_id from the admin dashboard
2. Set your WEBHOOK_LEAD_KEY in Railway environment variables
3. Use a tool like Postman or curl to test the endpoint
4. Check the leads table in your database to confirm creation

## Troubleshooting

### Common Issues

1. **401 Unauthorized**
   - Check your webhook_key matches the environment variable
   - Ensure WEBHOOK_LEAD_KEY is set in Railway

2. **404 User Not Found**
   - Verify the user_id exists in your database
   - Use the correct UUID format

3. **400 Validation Error**
   - Ensure target_status is either "prospect" or "customer"
   - Phone and name fields are required

4. **500 Server Error**
   - Check Railway logs for database connection issues
   - Verify the leads table structure matches the model

## Notes

- Leads created via webhook will have `source` set to "webhook" by default if not specified
- The `trigger` field can contain comma-separated values for multiple sequence triggers
- The `notes` field maps to the `journey` column in the database
- All timestamps are stored in UTC
