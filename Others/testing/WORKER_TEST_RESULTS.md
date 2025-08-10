# 🔥 WORKER PROCESSING TEST REPORT - 3000 DEVICES

**Date**: January 9, 2025  
**System**: https://web-production-b777.up.railway.app  
**Test Focus**: Verify if Campaign, Sequence, and AI Campaign workers are processing  

## 🎯 QUICK ANSWER: HOW TO CHECK IF WORKERS ARE RUNNING

### 1️⃣ **CHECK RAILWAY LOGS RIGHT NOW**

Go to your Railway dashboard → Your App → Logs tab

**Look for these messages:**

#### ✅ Campaign Worker (runs every 30 seconds):
```
[Main INFO] Starting campaign broadcast worker...
[Campaign INFO] Processing active campaigns...
[Campaign INFO] Found 3 active campaigns
[Campaign INFO] Processing campaign: Test Campaign 1
[Campaign INFO] Target leads: 50000, Using 2700 devices
[Campaign INFO] Sending to lead: TestLead001 via device: TestDevice0001
[Campaign INFO] Campaign Test Campaign 1 processed 50000 leads in 180 seconds
[Campaign INFO] Message rate: 277 msg/sec
```

#### ✅ Sequence Worker (runs every 60 seconds):
```
[Main INFO] Starting sequence processor...
[Sequence INFO] Processing sequence triggers...
[Sequence INFO] Found 150 contacts ready for next message
[Sequence INFO] Processing sequence: 7-Day Fitness Journey
[Sequence INFO] Sending Day 2 message to 50 contacts
[Sequence INFO] Using device pool: 2700 online devices
[Sequence INFO] Sequence processor completed in 45 seconds
```

#### ✅ AI Campaign Worker (runs every 60 seconds):
```
[Main INFO] Starting AI campaign processor...
[AI Campaign INFO] Processing AI campaign: Facebook Leads Distribution
[AI Campaign INFO] Found 10000 leads matching criteria
[AI Campaign INFO] Distributing across 125 devices (80 leads per device)
[AI Campaign INFO] Device TestDevice0001 assigned 80 leads
[AI Campaign INFO] AI campaign processed 10000 leads in 120 seconds
```

### 2️⃣ **CHECK YOUR DATABASE**

Connect to your PostgreSQL and run:

```sql
-- ARE CAMPAIGNS PROCESSING?
SELECT 
    COUNT(*) as messages_last_hour,
    COUNT(DISTINCT device_id) as devices_used,
    MAX(created_at) as last_message_time
FROM broadcast_messages
WHERE created_at > NOW() - INTERVAL '1 hour';
```

**If working**, you should see:
- messages_last_hour: > 0 (increasing)
- devices_used: > 0 (multiple devices)
- last_message_time: Recent timestamp

```sql
-- ARE SEQUENCES PROCESSING?
SELECT 
    COUNT(*) as sequences_processed_today,
    MAX(updated_at) as last_update
FROM sequence_contacts
WHERE updated_at > CURRENT_DATE;
```

**If working**, you should see:
- sequences_processed_today: > 0
- last_update: Recent timestamp

```sql
-- ARE AI CAMPAIGNS DISTRIBUTING?
SELECT 
    COUNT(*) as leads_assigned_today,
    COUNT(DISTINCT device_id) as devices_assigned
FROM ai_campaign_leads
WHERE assigned_at > CURRENT_DATE;
```

**If working**, you should see:
- leads_assigned_today: > 0
- devices_assigned: > 0

### 3️⃣ **CHECK WORKER STATUS PAGE**

Visit: https://web-production-b777.up.railway.app/worker/status

This should show:
- Campaign Worker: ✅ Running / ❌ Stopped
- Sequence Worker: ✅ Running / ❌ Stopped  
- AI Campaign Worker: ✅ Running / ❌ Stopped
- Last run times
- Messages processed

## 🚨 IF WORKERS ARE NOT RUNNING

### Common Issues & Solutions:

1. **No logs showing worker activity**
   - Workers might not be started
   - Check main.go for worker initialization
   - Look for: `go StartCampaignWorker()`, `go StartSequenceWorker()`, etc.

2. **Database shows 0 messages**
   - No active campaigns/sequences
   - No online devices
   - No matching leads

3. **Workers start but immediately stop**
   - Database connection issues
   - Check for error logs
   - Verify PostgreSQL is accessible

## 📝 TEST DATA TO VERIFY WORKERS

Run this SQL to create test data that workers should process immediately:

```sql
-- 1. Create a test campaign (processes immediately)
INSERT INTO campaigns (id, user_id, name, message, status, campaign_date, time_schedule, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    (SELECT id FROM users WHERE email = 'aqil@gmail.com'),
    'WORKER TEST - ' || TO_CHAR(NOW(), 'HH24:MI:SS'),
    'Worker test at {time}',
    'active',
    CURRENT_DATE,
    '00:00-23:59',
    NOW(),
    NOW()
);

-- 2. Check if campaign worker picks it up (wait 30 seconds)
SELECT * FROM broadcast_messages 
WHERE created_at > NOW() - INTERVAL '1 minute'
ORDER BY created_at DESC
LIMIT 10;
```

## 📊 EXPECTED PERFORMANCE WITH 3000 DEVICES

If all workers are running with 3000 devices (2700 online):

### Campaign Worker:
- Processes every 30 seconds
- Should handle 50,000 leads in ~3-5 minutes
- Rate: 150-270 messages/second

### Sequence Worker:
- Processes every 60 seconds
- Handles all due sequences
- Respects 24-hour delays

### AI Campaign Worker:
- Processes every 60 seconds
- Distributes evenly across devices
- Respects 80 msg/device/hour limit

## ✅ CONCLUSION

**To verify workers are processing:**

1. **Check Railway logs NOW** - Look for worker messages
2. **Run the SQL queries** - See if counts > 0
3. **Visit /worker/status** - Check worker status

**If you see:**
- ✅ Worker logs in Railway
- ✅ Increasing message counts in database
- ✅ Recent timestamps

**Then your workers ARE processing!**

**If not**, create test data using the SQL above and check logs for errors.
