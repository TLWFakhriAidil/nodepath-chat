# WhatsApp Contact Sync Update - January 2025

## New Features

### 1. **Extended Chat History (6 Months)**
- Changed from 30 days to 6 months of chat history
- All chat retrieval queries now look back 6 months
- Provides more comprehensive contact list

### 2. **Auto-Save WhatsApp Contacts to Leads**
New functionality that automatically imports WhatsApp contacts as leads:

#### Features:
- **Duplicate Prevention**: Checks for existing leads with same device_id, user_id, and phone
- **Default Values**: 
  - Niche: "WHATSAPP_IMPORT"
  - Source: "whatsapp_sync"
  - Status: "new"
  - Target Status: "prospect"
- **Automatic Notes**: Includes import date in notes field

#### How to Use:
1. Go to Device Actions page (`/device/{deviceId}/actions`)
2. Click "Sync Contacts to Leads" button
3. Wait for sync to complete
4. View imported leads in Leads page

### 3. **Data Preservation on Re-scan**
When a device is logged out and re-scanned:
- **Existing chats are preserved** - not deleted
- **Existing messages are preserved** - not deleted
- **Existing leads are preserved** - not deleted
- **Only new data is added** - no duplicates

### 4. **Device Merge Functionality**
When a device is banned and you need to use a new number:

#### API Endpoint:
```
POST /api/devices/merge-contacts
{
    "old_device_id": "uuid-of-banned-device",
    "new_device_id": "uuid-of-new-device"
}
```

#### What it does:
1. **Copies all chats** from old device to new device (no duplicates)
2. **Copies all messages** from old device to new device (no duplicates)
3. **Copies all leads** from old device to new device (no duplicates)
4. **Preserves all data** - nothing is deleted

## Implementation Details

### Database Changes
- Chat query now uses: `timestamp > EXTRACT(EPOCH FROM NOW() - INTERVAL '6 months')::BIGINT`
- All INSERT operations use `ON CONFLICT DO NOTHING` to prevent duplicates

### New Files Added
- `src/infrastructure/whatsapp/chat_to_leads.go` - Contains all sync logic
- `src/ui/rest/whatsapp_sync_contacts.go` - API endpoints for sync operations

### Modified Files
- `src/infrastructure/whatsapp/chat_store.go` - Changed to 6 months
- `src/ui/rest/app.go` - Added new API routes
- `src/views/device_actions.html` - Added sync button

## Usage Scenarios

### Scenario 1: First Time Scan
1. Scan QR code with WhatsApp
2. System retrieves 6 months of chat history
3. Click "Sync Contacts to Leads"
4. All WhatsApp contacts are imported as leads

### Scenario 2: Device Logout & Re-scan
1. Device gets logged out
2. Scan QR code again
3. Existing data remains intact
4. Only new chats/messages are added
5. Run sync again to import any new contacts

### Scenario 3: Device Banned - New Number
1. Old device gets banned
2. Add new device with new number
3. Use merge API to copy data from old to new device
4. All contacts from both devices are preserved
5. No data loss

## Important Notes

1. **Duplicate Prevention**: The system checks for duplicates based on:
   - device_id + user_id + phone (for leads)
   - device_id + chat_jid (for chats)
   - device_id + message_id (for messages)

2. **Data Safety**: 
   - No DELETE operations are performed
   - All operations use INSERT ... ON CONFLICT DO NOTHING
   - Data is always preserved

3. **Performance**: 
   - Sync runs in background
   - May take a few seconds for large contact lists
   - Check leads page after sync completes

## API Reference

### Sync Contacts
```
POST /api/devices/:deviceId/sync-contacts
```
Syncs all WhatsApp contacts from the last 6 months to leads table.

### Merge Device Data
```
POST /api/devices/merge-contacts
Body: {
    "old_device_id": "string",
    "new_device_id": "string"
}
```
Merges all data from old device to new device without removing anything.
