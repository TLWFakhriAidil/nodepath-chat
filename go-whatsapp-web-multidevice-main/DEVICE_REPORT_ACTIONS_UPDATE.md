# Device Report Actions & Real-time Status Update

## Date: January 2025

## Features Implemented:

### 1. Device Report Actions Column
Added a new "Actions" column in the device report with two icons:

#### a) Retry Icon (🔄)
- **Purpose**: Retry failed leads
- **Functionality**: Already existed, allows retrying failed lead deliveries
- **Applies to**: All failed leads regardless of campaign type

#### b) Transfer Icon (📤)
- **Purpose**: Transfer successful AI campaign leads to regular leads table
- **Functionality**: 
  - Only visible for AI campaigns (campaign_type = 'ai')
  - Only enabled when device status is 'active' or 'connected'
  - Only for successful leads (status = 'delivered')
  - Creates a duplicate entry in the regular `leads` table
  - Keeps the original in `leads_ai` table
  - Allows device to run personal campaigns with these leads

### 2. API Endpoint for Lead Transfer
Created `/api/ai-campaign/transfer-leads` endpoint that:
- Validates the device belongs to the user
- Checks if device is active/connected
- Transfers only successful AI campaign leads
- Duplicates leads from `leads_ai` to `leads` table
- Maintains data integrity

### 3. Real-time Device Status
Implemented real-time device connection status checking:

#### a) New Endpoint
- `/api/devices/check-connection` - Checks actual WhatsApp connection status
- Updates database with current connection state
- Returns real-time status for all user devices

#### b) Dashboard Integration
- `loadDevices()` now checks real-time status before fetching device list
- Ensures device status shown matches actual WhatsApp connection
- Auto-refresh (if enabled) will update status every 10 seconds

#### c) Logout Status Update
- When device is logged out, status is properly set to "offline"
- Clear session endpoint updates device status in database
- Prevents showing "connected" status after logout

## Files Modified:

1. **src/views/dashboard.html**
   - Added Actions column to device report table
   - Added `transferAILeadsToDevice()` function
   - Updated `loadDevices()` to check real-time status
   - Modified `displayDeviceReport()` to show action buttons

2. **src/ui/rest/transfer_ai_leads.go** (New)
   - Created endpoint to handle AI lead transfers
   - Validates permissions and device status
   - Duplicates leads from leads_ai to leads table

3. **src/ui/rest/check_device_connection.go** (New)
   - Created endpoint to check real-time connection status
   - Updates database with current status
   - Returns status for all user devices

4. **src/ui/rest/app.go**
   - Added routes for new endpoints
   - `/api/ai-campaign/transfer-leads`
   - `/api/devices/check-connection`

## How It Works:

### Lead Transfer Process:
1. User clicks transfer icon in device report
2. System confirms the action
3. API validates device ownership and status
4. Successful AI leads are duplicated to regular leads table
5. Device can now use these leads for personal campaigns

### Real-time Status:
1. Dashboard loads or refreshes
2. First calls check-connection endpoint
3. Server checks actual WhatsApp client status
4. Updates database if status changed
5. Returns current status to dashboard
6. Dashboard displays accurate connection status

## Benefits:
- AI campaign leads can be reused for personal campaigns
- Device status accurately reflects WhatsApp connection
- No more stale "connected" status after logout
- Better user experience with real-time updates
