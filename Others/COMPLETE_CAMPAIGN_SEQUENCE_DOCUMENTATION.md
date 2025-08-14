# Complete Campaign & Sequence System Documentation

## 🎯 Overview

This WhatsApp Multi-Device System supports two types of messaging:
1. **Campaigns** - One-time broadcast to all matching leads
2. **Sequences** - Multi-step drip campaigns with trigger-based progression

Both use the SAME infrastructure but differ in scheduling and progression logic.

## 📊 Campaign System

### What is a Campaign?
A one-time message blast sent to all leads matching specific criteria (niche + target_status) on a scheduled date/time.

### Campaign Features:
- **One-time execution** - Runs once on scheduled date
- **Bulk messaging** - Sends to all matching leads simultaneously  
- **Device-specific** - Each device only sends to its own leads
- **Target filtering** - By niche (e.g., "fitness") and status (prospect/customer/all)
- **Time scheduling** - Specific date + time window
- **Media support** - Text or image messages
- **Anti-spam** - Malaysian greetings + message randomization
- **Status tracking** - scheduled → triggered → completed/failed

### Campaign Database Schema:
```sql
campaigns:
- id (int)
- user_id 
- title
- message
- image_url
- niche
- target_status (prospect/customer/all)
- campaign_date
- time_schedule
- min_delay_seconds
- max_delay_seconds
- status
```

### Campaign Process Flow:
```
1. Admin creates campaign with:
   - Message content
   - Target niche + status
   - Schedule date/time
   - Min/max delays

2. Campaign Trigger Service (runs every minute):
   - Finds campaigns where date = today AND time = now (±10 min)
   - Gets all user's online devices
   - For EACH device:
     - Get leads matching (device_id + niche + status)
     - Create BroadcastMessage for each lead
     - Queue to database

3. Broadcast Worker (runs every 5 seconds):
   - Picks messages from database queue
   - Routes to appropriate device worker
   - Updates status: pending → processing

4. Device Worker:
   - Applies human-like delay (random between min/max)
   - Sends to WhatsAppMessageSender

5. WhatsAppMessageSender:
   - Checks if platform device (Wablas/Whacenter) or WhatsApp Web
   - Applies anti-spam (greeting + randomization)
   - Sends message
   - Updates status: processing → sent/failed
```

## 🔄 Sequence System

### What is a Sequence?
A multi-step automated message series where each contact progresses through steps based on triggers and delays.

### Sequence Features:
- **Multi-step flow** - Up to 30+ steps/days
- **Trigger-based progression** - Each step has entry/exit triggers
- **Individual timelines** - Each contact progresses independently
- **Flexible delays** - Hours between steps (trigger_delay_hours)
- **Entry points** - Multiple ways to enter sequence
- **Sequence chaining** - Can link to other sequences
- **Same anti-spam** - Malaysian greetings + randomization
- **Per-step customization** - Different delays per step

### Sequence Database Schema:
```sql
sequences:
- id
- user_id
- name
- niche
- trigger (main sequence trigger)
- target_status
- status (active/inactive)
- min_delay_seconds (fallback)
- max_delay_seconds (fallback)

sequence_steps:
- id
- sequence_id
- day_number
- trigger (e.g., "fitness_day1")
- next_trigger (e.g., "fitness_day2")
- trigger_delay_hours (hours to wait)
- is_entry_point (boolean)
- content (message text)
- message_type (text/image)
- media_url
- min_delay_seconds (per-step delays)
- max_delay_seconds (per-step delays)

sequence_contacts:
- id
- sequence_id
- contact_phone
- contact_name
- current_step
- current_trigger
- next_trigger_time
- status (active/pending/sent/failed)
- assigned_device_id
- processing_device_id
- sequence_stepid

leads:
- trigger (comma-separated, e.g., "fitness_start,nutrition_basic")
```

### Sequence Process Flow:
```
1. Lead gets trigger assigned:
   UPDATE leads SET trigger = 'fitness_start'

2. Sequence Trigger Processor (runs every 15 seconds):
   
   A. ENROLLMENT PHASE:
      - Find leads with triggers matching sequence entry points
      - Create sequence_contacts records for ALL steps
      - First step = 'active', others = 'pending'
      - Calculate next_trigger_time for each step
      - Assign device_id from lead

   B. PROCESSING PHASE:
      - Find active contacts where next_trigger_time <= NOW
      - Only process if assigned device is online
      - Create BroadcastMessage with:
        * UserID, DeviceID, SequenceID
        * RecipientName (for greeting)
        * Min/max delays FROM STEP (not sequence)
      - Queue to database (same as campaigns)

3. After message sent:
   - Mark current step as 'sent'
   - Activate next step (if exists)
   - Or remove trigger from lead (if sequence complete)

4. Broadcast flow identical to campaigns (steps 3-5 above)
```

## 🔍 Key Differences

| Feature | Campaigns | Sequences |
|---------|-----------|-----------|
| **Execution** | One-time on date | Multi-step over days/weeks |
| **Progression** | All at once | Individual timeline per contact |
| **Trigger System** | No | Yes - trigger-based flow |
| **Entry Method** | Date/time match | Lead trigger match |
| **Delays Between** | No | Yes - trigger_delay_hours |
| **Completion** | When all sent | When last step reached |
| **Chaining** | No | Yes - to other sequences |

## 🛡️ Shared Features (Both Systems)

### 1. **Device Assignment**
- Leads belong to specific devices
- Only the owning device can send to that lead
- No cross-device messaging
- Platform devices (Wablas/Whacenter) supported

### 2. **Anti-Spam System**
- **Malaysian Greetings**: "Hi Cik, apa khabar"
- **Message Randomization**: Homoglyphs + zero-width chars
- **Human Delays**: Random between min/max seconds
- **Applied at send layer** (not in campaign/sequence logic)

### 3. **Message Queue**
- Both queue to `broadcast_messages` table
- Same fields: UserID, DeviceID, RecipientName, delays, etc.
- Same status flow: pending → processing → sent/failed
- Same broadcast worker processing

### 4. **Platform Support**
```go
if device.Platform != "" {
    // Send via Wablas/Whacenter API
    // With anti-spam applied
} else {
    // Send via WhatsApp Web
    // With anti-spam applied
}
```

## 📋 Implementation Details

### Broadcast Message Structure (Shared):
```go
BroadcastMessage{
    UserID:         string       // Owner
    DeviceID:       string       // Assigned device
    CampaignID:     *int         // If from campaign
    SequenceID:     *string      // If from sequence
    RecipientPhone: string
    RecipientName:  string       // For greetings
    Message:        string
    Type:           string       // text/image
    MediaURL:       string
    MinDelay:       int          // Seconds
    MaxDelay:       int          // Seconds
    ScheduledAt:    time.Time
    Status:         string       // pending/processing/sent/failed
}
```

### Anti-Spam Applied (Both):
```go
// 1. Add greeting
messageWithGreeting = greetingProcessor.PrepareMessageWithGreeting(
    originalMessage,
    recipientName,
    deviceID,
    phone
)

// 2. Randomize
finalMessage = messageRandomizer.RandomizeMessage(messageWithGreeting)

// Result: "Hi Cik, apa khabar\n\nYоur mеssage with invisible chars"
```

## ✅ System Capabilities

- **Scale**: 3000+ devices simultaneously
- **Volume**: 15,000-20,000 messages/hour (safe rate)
- **Reliability**: Database queue with retry capability
- **Tracking**: Full analytics and reporting
- **Flexibility**: Campaigns for blasts, Sequences for nurturing

## 🚀 Usage Examples

### Campaign Example:
```
"Black Friday Sale" campaign
- Target: niche="ecommerce", status="customer"
- Date: 2024-11-29, Time: 10:00 AM
- Message: "50% off everything today only!"
- Result: Sends to all customers interested in ecommerce at 10 AM
```

### Sequence Example:
```
"30 Day Fitness Challenge" sequence
- Entry: Lead gets trigger "fitness_start"
- Day 1: Welcome message (immediate)
- Day 2: First workout (24 hours later)
- Day 3: Nutrition tips (24 hours later)
- ...
- Day 30: Congratulations (24 hours later)
- Result: Each lead progresses individually through 30 days
```

## 🎯 Summary

Both campaigns and sequences:
1. Use the same infrastructure
2. Queue messages to database
3. Process through same workers
4. Apply same anti-spam
5. Respect device ownership

The ONLY difference is:
- **Campaigns**: One-time scheduled blast
- **Sequences**: Multi-step trigger-based progression

Everything else is identical - making the system reliable, scalable, and maintainable.