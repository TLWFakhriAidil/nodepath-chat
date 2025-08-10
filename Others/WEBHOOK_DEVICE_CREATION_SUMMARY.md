# Webhook Device Creation Summary

## How Device Creation Works

When the webhook receives a request and the device doesn't exist, it creates a new device with these exact mappings:

### Field Mappings:

| user_devices Column | Value Source | Example |
|-------------------|--------------|---------|
| **id** | `device_id` from request | "abc123-def456-ghi789" |
| **user_id** | `user_id` from request | "xyz987-wvu654-tsr321" |
| **device_name** | `device_name` from request | "ADHQ-S13" |
| **phone** | null (empty string) | "" |
| **jid** | `device_id` from request | "abc123-def456-ghi789" |
| **status** | 'online' (hardcoded) | "online" |
| **last_seen** | current timestamp | 2025-01-15 10:30:00 |
| **created_at** | current timestamp | 2025-01-15 10:30:00 |
| **updated_at** | current timestamp | 2025-01-15 10:30:00 |
| **min_delay_seconds** | 5 (hardcoded) | 5 |
| **max_delay_seconds** | 15 (hardcoded) | 15 |
| **platform** | `platform` from request | "Whacenter" |

## Important Notes:

1. **ID = JID**: Both `id` and `jid` columns store the same `device_id` value
2. **Phone is NULL**: The phone column is left empty (empty string)
3. **Status is Online**: All devices created via webhook start with 'online' status
4. **Default Delays**: Min 5 seconds, Max 15 seconds for all webhook-created devices

## Example Request:

```json
{
  "name": "Real Customer",
  "phone": "60198765432",
  "target_status": "customer",
  "device_id": "abc123-def456-ghi789",
  "user_id": "xyz987-wvu654-tsr321",
  "device_name": "ADHQ-S13",
  "platform": "Whacenter",
  "niche": "FITNESS",
  "trigger": "welcome_sequence"
}
```

## What Happens:

1. **Check Device**: System checks if device with id "abc123-def456-ghi789" exists
2. **Create Device** (if not exists):
   - id = "abc123-def456-ghi789"
   - jid = "abc123-def456-ghi789" (same value)
   - device_name = "ADHQ-S13"
   - platform = "Whacenter"
   - status = "online"
   - phone = "" (empty)
3. **Create Lead**: Lead is created with all provided data including platform

## SQL Migration Required:

```sql
-- Add platform column if not exists
ALTER TABLE user_devices 
ADD COLUMN IF NOT EXISTS platform VARCHAR(255);

ALTER TABLE leads
ADD COLUMN IF NOT EXISTS platform VARCHAR(255);
```

## Testing:

Use this exact format in Postman or your integration:

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

This will create a device where both `id` and `jid` = "test-device-123"
