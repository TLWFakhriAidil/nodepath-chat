# Webhook Device ID Logic

## Simple Rules:

### 1. For Non-UUID device_id (e.g., long strings):
```
device_id: hulN3t1yRMe2J48NYABNss3BOcYRMY1L9UsxdBLDGvkHlR53pBMYCYW.etRzghNw
```
**Result:**
- `id` = New generated UUID
- `jid` = Full original (hulN3t1yRMe2J48NYABNss3BOcYRMY1L9UsxdBLDGvkHlR53pBMYCYW.etRzghNw)
- Device name includes first 6 chars: "Device-hulN3t"

### 2. For Valid UUID device_id:
```
device_id: 22f6f5bd-56a5-4f1e-ac78-d4f33aa75158
```
**Result:**
- `id` = Same UUID (22f6f5bd-56a5-4f1e-ac78-d4f33aa75158)
- `jid` = Same UUID (22f6f5bd-56a5-4f1e-ac78-d4f33aa75158)

## Examples:

### Example 1: Non-UUID Device ID
**Request:**
```json
{
  "device_id": "hulN3t1yRMe2J48NYABNss3BOcYRMY1L9UsxdBLDGvkHlR53pBMYCYW.etRzghNw",
  "user_id": "de078f16-3266-4ab3-8153-a248b015228f",
  "device_name": "SCVTC-S21",
  // ... other fields
}
```

**Database Result:**
- `id`: `abc123-generated-uuid` (new UUID)
- `jid`: `hulN3t1yRMe2J48NYABNss3BOcYRMY1L9UsxdBLDGvkHlR53pBMYCYW.etRzghNw` (full original)

### Example 2: UUID Device ID
**Request:**
```json
{
  "device_id": "22f6f5bd-56a5-4f1e-ac78-d4f33aa75158",
  "user_id": "de078f16-3266-4ab3-8153-a248b015228f",
  "device_name": "ADHQ-S13",
  // ... other fields
}
```

**Database Result:**
- `id`: `22f6f5bd-56a5-4f1e-ac78-d4f33aa75158` (same)
- `jid`: `22f6f5bd-56a5-4f1e-ac78-d4f33aa75158` (same)

## Summary:
- **Non-UUID**: First 6 chars used as prefix, new UUID generated for ID, full string saved in JID
- **Valid UUID**: Used as-is for both ID and JID
- **JID always contains the full original device_id**
