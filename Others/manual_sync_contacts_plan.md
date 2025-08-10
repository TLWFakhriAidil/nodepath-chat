--- Manual Sync Contacts Implementation Plan ---

## Current Behavior
The system automatically syncs WhatsApp contacts to leads after:
1. History sync event (when WhatsApp loads chat history)
2. This happens in `chat_store.go` line 191: `AutoSaveChatsToLeads(deviceID, device.UserID)`

## Changes Required

### 1. Remove Auto-Sync from History Handler
In `src/infrastructure/whatsapp/chat_store.go`:
- Comment out or remove the auto-sync call after history sync (lines 183-195)

### 2. API Endpoint Already Exists
Good news! The manual sync API endpoint already exists:
- Endpoint: `POST /api/devices/:id/sync-contacts`
- Handler: `SyncWhatsAppContacts` in `whatsapp_sync_contacts.go`
- It calls `AutoSaveChatsToLeads(deviceId, session.UserID)` in background

### 3. Add UI Button
Need to add a "Sync Contacts" button in the device actions page that calls this endpoint.

### 4. Testing
The sync function saves contacts from the last 6 months to the leads table in MySQL.