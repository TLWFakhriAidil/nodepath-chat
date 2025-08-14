# AI Campaign Implementation Guide

## Overview
This guide provides step-by-step instructions for implementing the AI Campaign feature in your WhatsApp Multi-Device System.

## Implementation Steps

### 1. Database Migration
First, run the database migration to create new tables and add columns:

```bash
# Run the migration SQL file
psql $DATABASE_URL < ai_campaign_implementation/001_ai_campaign_migration.sql
```

### 2. Copy Model Files
Copy the new model to the models directory:

```bash
# Copy LeadAI model
cp ai_campaign_implementation/lead_ai.go src/models/lead_ai.go
```

### 3. Update Campaign Model
Add the new fields to `src/models/campaign.go`:

```go
// Add these fields to the Campaign struct
AI    *string `json:"ai" db:"ai"`              // "ai" for AI campaigns, null for regular
Limit int     `json:"limit" db:"\"limit\""`     // Device limit for AI campaigns
```

### 4. Add Repository
Copy the AI lead repository:

```bash
cp ai_campaign_implementation/lead_ai_repository.go src/repository/lead_ai_repository.go
```

### 5. Add AI Campaign Processor
Copy the processor to the usecase directory:

```bash
cp ai_campaign_implementation/ai_campaign_processor.go src/usecase/ai_campaign_processor.go
```

### 6. Update API Handlers
Add the new handler functions from `ai_lead_handlers.go` to `src/ui/rest/app.go`.

### 7. Modify CreateCampaign Function
Replace the existing `CreateCampaign` function in `src/ui/rest/app.go` with the modified version from `create_campaign_modified.go`.

### 8. Add Trigger Handler
Add the `TriggerAICampaign` function from `trigger_ai_campaign.go` to `src/ui/rest/app.go`.

### 9. Update Routes
Add the new routes from `routes_to_add.go` to `src/cmd/rest.go`.

### 10. Update Dashboard HTML
Add the Manage AI tab from `manage_ai_tab.html` to `src/views/dashboard.html`:
- Add the tab navigation item after existing tabs
- Add the tab content after existing tab contents
- Add the modals at the end of the file

### 11. Add JavaScript Functions
Add the JavaScript from `ai_management.js` to your dashboard JavaScript file.

### 12. Update Campaign Display
Replace the `displayCampaigns` function with the modified version from `campaign_display_modified.js`.

### 13. Update Broadcast Processing
Modify your existing broadcast processor to check for AI campaigns and route them to the AI processor.

## Testing the Implementation

### 1. Create Test AI Leads
1. Navigate to the "Manage AI" tab
2. Click "Add AI Lead"
3. Create 10-20 test leads with:
   - Same niche value
   - Mix of prospect/customer target status
   - Valid phone numbers

### 2. Create AI Campaign
1. Click "Create AI Campaign"
2. Set:
   - Matching niche value
   - Device limit (e.g., 5 per device)
   - Target status (all/prospect/customer)
   - Message content

### 3. Connect Test Devices
1. Ensure you have at least 2-3 connected WhatsApp devices
2. Check device status in the Devices tab

### 4. Trigger Campaign
1. Find the AI campaign in the campaigns list (marked with robot icon)
2. Click "Trigger" button
3. Monitor the logs for round-robin assignment

### 5. Verify Results
1. Check AI leads table - status should update from "pending" to "sent" or "failed"
2. Check device_id assignment in the leads
3. Verify messages are being sent via WhatsApp
4. Check that no device exceeds the set limit

## Troubleshooting

### Common Issues

1. **"No connected devices" error**
   - Ensure devices are connected and online
   - Check device status in the database

2. **Leads not being assigned**
   - Verify leads have matching niche
   - Check target_status matches campaign settings
   - Ensure leads have "pending" status

3. **Campaign stuck in "triggered" status**
   - Check logs for errors
   - Verify Redis is running
   - Check broadcast workers are active

### Debug Queries

```sql
-- Check AI leads status
SELECT status, COUNT(*) FROM leads_ai GROUP BY status;

-- Check campaign progress
SELECT * FROM ai_campaign_progress WHERE campaign_id = ?;

-- Check device assignments
SELECT device_id, COUNT(*) FROM leads_ai 
WHERE device_id IS NOT NULL 
GROUP BY device_id;
```

## Performance Considerations

1. **Concurrent Processing**: The system processes 5 leads concurrently by default
2. **Delays**: Respects min/max delay settings between messages
3. **Device Limits**: Strictly enforces per-device limits
4. **Failure Handling**: Marks device as failed after 3 consecutive errors

## Future Enhancements

1. **Real-time Progress**: WebSocket updates for live campaign monitoring
2. **Retry Mechanism**: Automatic retry for failed leads after cooldown
3. **Priority Queue**: High-priority leads processed first
4. **Analytics Dashboard**: Detailed campaign performance metrics
5. **Bulk Import**: CSV import for AI leads