# Webhook UUID Handling for Non-UUID Device IDs

## How It Works

When you send a device_id like:
```
hulN3t1yRMe2J48NYABNss3BOcYRMY1L9UsxdBLDGvkHlR53pBMYCYW.etRzghNw
```

The webhook will:

1. **Check if it's a valid UUID**
   - If YES → Use as is
   - If NO → Generate a new UUID

2. **For non-UUID device_ids:**
   - Takes first 6 characters: `hulN3t`
   - Generates a new UUID: `abc123-def456-ghi789...`
   - Saves in database as:
     - `id` = Generated UUID
     - `jid` = Full original device_id
     - `device_name` = "Device-hulN3t" (if not provided)

## Example

### Request:
```json
{
  "name": "John Doe",
  "phone": "60123456789",
  "target_status": "prospect",
  "device_id": "hulN3t1yRMe2J48NYABNss3BOcYRMY1L9UsxdBLDGvkHlR53pBMYCYW.etRzghNw",
  "user_id": "de078f16-3266-4ab3-8153-a248b015228f",
  "device_name": "SCVTC-S21",
  "platform": "Whacenter",
  "niche": "EXSTART",
  "trigger": "NEWNP"
}
```

### What Gets Saved:

**user_devices table:**
- `id` = `abc123-def456-ghi789...` (Generated UUID)
- `jid` = `hulN3t1yRMe2J48NYABNss3BOcYRMY1L9UsxdBLDGvkHlR53pBMYCYW.etRzghNw` (Full original)
- `device_name` = `SCVTC-S21`
- `user_id` = `de078f16-3266-4ab3-8153-a248b015228f`
- `platform` = `Whacenter`

**leads table:**
- `device_id` = `abc123-def456-ghi789...` (The UUID from user_devices.id)

### Response:
```json
{
  "status": 200,
  "code": "SUCCESS",
  "message": "Lead created successfully",
  "results": {
    "lead_id": "generated-uuid",
    "device_id": "abc123-def456-ghi789...",  // The actual UUID used
    "device_jid": "hulN3t1yRMe2J48NYABNss3BOcYRMY1L9UsxdBLDGvkHlR53pBMYCYW.etRzghNw", // Original
    "device_created": true,
    // ... other fields
  }
}
```

## Key Points

1. **ID must be UUID**: PostgreSQL requires UUID format for the id column
2. **JID stores original**: The full original device_id is preserved in JID column
3. **Automatic handling**: You don't need to worry about UUID format - just send your device_id
4. **Duplicate prevention**: Checks by user_id + jid to prevent duplicates
5. **Tracking**: You can always find your device by searching JID column with original device_id

## Benefits

- ✅ Accepts any device_id format
- ✅ No UUID syntax errors
- ✅ Preserves full original device_id
- ✅ Automatic UUID generation
- ✅ Works with existing UUID constraint
