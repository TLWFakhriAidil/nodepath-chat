# CAMPAIGN FLOW ANALYSIS

## 🔍 COMPLETE CAMPAIGN FLOW (Step by Step):

### 1. **Campaign Creation**
- Admin creates campaign with niche (GRR), target_status (prospect), message, and schedule time
- Campaign saved in database with status = 'pending'

### 2. **Campaign Trigger Service** (runs every 60 seconds - found in config)
The `OptimizedCampaignTrigger.ProcessCampaigns()` does:

```go
// Finds campaigns where:
// - status = 'pending' 
// - scheduled time <= NOW
```

### 3. **Campaign Execution** (in executeCampaign function)
When a campaign is found:

a) **Get User Devices** 
   - Gets all devices for the user
   - Filters only "online" or "connected" devices
   - Platform devices (Wablas/Whacenter) always considered online

b) **Get Matching Leads**
   - For EACH online device:
   - Get leads where device_id = device AND niche = 'GRR' AND target_status = 'prospect'
   - Combines all leads from all devices

c) **Create Broadcast Messages**
   - For EACH lead found:
   ```go
   msg := BroadcastMessage{
       UserID: campaign.UserID,
       DeviceID: lead.DeviceID,
       CampaignID: &campaign.ID,
       RecipientPhone: lead.Phone,
       Content: campaign.Message,
       ScheduledAt: time.Now(),
   }
   broadcastRepo.QueueMessage(msg)
   ```

d) **Update Campaign Status**
   - If messages created > 0: status = 'triggered'
   - If no messages: status = 'finished'

### 4. **Broadcast Processor** (runs every 2 seconds)
The `UltraOptimizedBroadcastProcessor` does:

```sql
SELECT FROM broadcast_messages 
WHERE status = 'pending' 
AND scheduled_at <= NOW()
LEFT JOIN campaigns to get min/max delays
```

Then:
- Creates broadcast pools for each campaign
- Queues messages to workers
- Workers send via WhatsApp

## 🚨 POTENTIAL ISSUES I SEE:

### 1. **Campaign Status Not Updated**
Your logs show campaigns 59 & 60 keep finding leads but status stays 'pending' or 'failed'. This suggests:
- The `executeCampaign` is running in a goroutine
- It might be failing BEFORE updating status
- The error might not be logged

### 2. **Missing Scheduler**
I can't find where `ProcessCampaigns()` is actually being called! The campaign trigger might not be running at all.

### 3. **The "ShouldSend: 1" Log**
This log comes from a different function that just COUNTS leads, not the actual campaign execution.

## 🔧 THE REAL PROBLEM:

**The campaign trigger scheduler might not be running!** I couldn't find where `ProcessCampaigns()` is scheduled to run every minute.

Without the scheduler, campaigns will never move from 'pending' to 'triggered' and no broadcast messages will be created.
