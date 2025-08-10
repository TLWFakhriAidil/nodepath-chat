# Clean Webhook Test Example

## What You Send:
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

## What Gets Saved in Database:

| Column | Value |
|--------|-------|
| id | auto-generated-uuid |
| name | John Doe |
| phone | 60123456789 |
| target_status | prospect |
| device_id | device-id |
| user_id | user_id |
| niche | EXSTART |
| trigger | NEWNP |
| created_at | 2025-01-15 10:30:00 |
| updated_at | 2025-01-15 10:30:00 |
| source | NULL |
| status | NULL |
| email | NULL |
| notes | NULL |

## Quick Test Command:
```bash
curl -X POST http://localhost:3000/webhook/lead/create \
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

## Response You'll Get:
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

## Key Points:
- ✅ Only saves exactly what you send
- ✅ No default values added
- ✅ No authentication required
- ✅ Direct field mapping
- ✅ Returns all saved data in response
