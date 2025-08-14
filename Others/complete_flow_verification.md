# COMPLETE A-Z VERIFICATION CHECKLIST

## 1. CAMPAIGN FLOW (A to Z)

### A. Campaign Creation
- File: `src/ui/rest/app.go` - CreateCampaign()
- Creates campaign record in campaigns table

### B. Campaign Triggering  
- File: `src/usecase/campaign_trigger.go` - ProcessCampaignTriggers()
- Runs every minute via StartTriggerProcessor()
- Gets active campaigns and matching leads
- Creates messages in broadcast_messages table

### C. Duplicate Prevention for Campaigns
- Location: `src/repository/broadcast_repository.go` - QueueMessage()
- Check: campaign_id + recipient_phone + device_id
- Must be unique combination

### D. Message Processing
- File: `src/usecase/optimized_broadcast_processor.go`
- Calls GetPendingMessagesAndLock() with worker ID
- Prevents multiple workers from getting same message

### E. Message Sending
- File: `src/infrastructure/broadcast/device_worker.go`
- Sends via WhatsApp
- Updates status to 'sent'

## 2. SEQUENCE FLOW (A to Z)

### A. Sequence Creation
- File: `src/ui/rest/app.go` - CreateSequence()
- Creates sequence with steps

### B. Sequence Enrollment
- File: `src/usecase/direct_broadcast_processor.go`
- Enrolls leads based on trigger
- Creates sequence_contacts record

### C. Daily Message Creation
- File: `src/usecase/campaign_trigger.go` - ProcessDailySequenceMessages()
- Runs daily for each enrolled contact
- Checks current day and creates next message

### D. Duplicate Prevention for Sequences
- Location: `src/repository/broadcast_repository.go` - QueueMessage()
- Check: sequence_stepid + recipient_phone + device_id
- Must be unique combination

### E. Message Processing & Sending
- Same as campaigns - uses GetPendingMessagesAndLock()

## CRITICAL CHECKS NEEDED:

1. ✅ Database columns exist (verified above)
2. ❌ Worker ID not being set (0% have worker IDs)
3. ✅ No duplicates in last 7 days
4. ❓ Need to verify GetPendingMessagesAndLock is called everywhere
5. ❓ Need to verify duplicate checks in QueueMessage
