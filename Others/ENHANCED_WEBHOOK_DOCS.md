# Enhanced Webhook with Auto Device Creation

## New Features
The webhook now supports automatic device creation and platform tracking!

## Updated Request Format

```json
{
  "name": "Real Customer",
  "phone": "60198765432",
  "target_status": "customer",
  "niche": "FITNESS",
  "trigger": "welcome_sequence",
  "device_id": "abc123-def456-ghi789",
  "user_id": "xyz987-wvu654-tsr321",
  "device_name": "ADHQ-S13",
  "platform": "Whacenter"
}
```

## How It Works

### 1. Device Check & Creation
When you send a webhook request, the system will:

1. **Check if device exists** by `device_id`
2. **If device doesn't exist**, it creates a new device with:
   - `id` = Your provided `device_id`
   - `user_id` = Your provided `user_id`
   - `device_name` = Your provided `device_name` (or auto-generated)
   - `phone` = empty (null)
   - `jid` = empty (null)
   - `status` = 'online'
   - `last_seen` = current timestamp
   - `created_at` = current timestamp
   - `updated_at` = current timestamp
   - `min_delay_seconds` = 5
   - `max_delay_seconds` = 15
   - `platform` = Your provided `platform`

### 2. Lead Creation
After ensuring the device exists, it creates the lead with all your data including the `platform` field.

## New Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| device_name | string | No | Name for the device (e.g., "ADHQ-S13") |
| platform | string | No | Platform name (e.g., "Whacenter") |

## Response

The response now includes a `device_created` flag:

```json
{
  "status": 200,
  "code": "SUCCESS",
  "message": "Lead created successfully",
  "results": {
    "lead_id": "generated-uuid",
    "name": "Real Customer",
    "phone": "60198765432",
    "niche": "FITNESS",
    "trigger": "welcome_sequence",
    "target_status": "customer",
    "device_id": "abc123-def456-ghi789",
    "user_id": "xyz987-wvu654-tsr321",
    "platform": "Whacenter",
    "device_created": true,  // Indicates if a new device was created
    "created_at": "2025-01-15T10:30:00Z"
  }
}
```

## Database Changes

### New Columns Added:
1. **user_devices.platform** - VARCHAR(255)
2. **leads.platform** - VARCHAR(255)

### Migration SQL:
```sql
-- Add platform column to user_devices table
ALTER TABLE user_devices 
ADD COLUMN IF NOT EXISTS platform VARCHAR(255);

-- Add platform column to leads table  
ALTER TABLE leads
ADD COLUMN IF NOT EXISTS platform VARCHAR(255);
```

## Example Use Cases

### 1. First Time Device & Lead
```json
{
  "name": "New Customer",
  "phone": "60123456789",
  "target_status": "prospect",
  "device_id": "new-device-001",
  "user_id": "your-user-id",
  "device_name": "Whacenter-Bot-1",
  "platform": "Whacenter",
  "niche": "EXSTART",
  "trigger": "NEWNP"
}
```
Result: Creates both device and lead

### 2. Existing Device, New Lead
```json
{
  "name": "Another Customer",
  "phone": "60198765432",
  "target_status": "customer",
  "device_id": "new-device-001",  // Same device ID as above
  "user_id": "your-user-id",
  "platform": "Whacenter",
  "niche": "FITNESS",
  "trigger": "welcome"
}
```
Result: Uses existing device, creates only lead

## Testing in Postman

1. **Method**: POST
2. **URL**: `https://web-production-b777.up.railway.app/webhook/lead/create`
3. **Headers**: 
   - `Content-Type: application/json`
4. **Body** (raw JSON):
```json
{
  "name": "Test Customer",
  "phone": "60123456789",
  "target_status": "prospect",
  "device_id": "test-device-123",
  "user_id": "your-actual-user-id",
  "device_name": "Test-Device",
  "platform": "Whacenter",
  "niche": "EXSTART",
  "trigger": "NEWNP"
}
```

## Benefits

1. **Automatic Device Management**: No need to manually create devices first
2. **Platform Tracking**: Track which platform/service created each lead
3. **Flexible Integration**: Works with existing devices or creates new ones
4. **Complete Audit Trail**: Know exactly where each lead came from

## Notes

- Device IDs should be unique across your system
- If device_name is not provided, system auto-generates one
- Platform field helps track lead sources (Whacenter, ManualEntry, API, etc.)
- All timestamps are in UTC
