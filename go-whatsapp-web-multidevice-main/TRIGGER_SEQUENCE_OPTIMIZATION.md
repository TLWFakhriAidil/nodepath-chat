# Trigger-Based Sequence System for 3000 Devices

## Overview
This document describes the optimized trigger-based sequence system designed to handle 3000 devices simultaneously with excellent performance.

## Architecture

### 1. Trigger System
- Leads have a `trigger` column that can contain multiple comma-separated triggers
- Example: `"fitness_start,crypto_welcome,realestate_intro"`
- Each trigger maps to a sequence step

### 2. Processing Flow

```
Lead.trigger = "fitness_start,crypto_start"
    ↓
Sequence Processor (every 30 seconds)
    ↓
1. Check active sequences with matching triggers
2. Enroll lead in sequence_contacts if not exists
3. Process sequence_contacts ready for next message
4. Distribute across 3000 devices
5. Update progress and move to next trigger
```

### 3. Database Structure

#### Leads Table
- `trigger` (VARCHAR 1000) - Comma-separated list of active triggers

#### Sequence Steps
- `trigger` (VARCHAR 255) - Unique trigger identifier  
- `next_trigger` (VARCHAR 255) - Points to next step
- `trigger_delay_hours` (INT) - Hours to wait before next trigger
- `is_entry_point` (BOOLEAN) - Marks sequence start points

#### Sequence Contacts
- `current_trigger` (VARCHAR 255) - Current position in sequence
- `next_trigger_time` (TIMESTAMP) - When to process next
- `processing_device_id` (UUID) - Prevents double processing
- `retry_count` (INT) - For error handling

### 4. How It Works

#### Step 1: Lead Assignment
```sql
-- Add trigger to lead
UPDATE leads SET trigger = 'fitness_start' WHERE phone = '60123456789';

-- Or multiple triggers
UPDATE leads SET trigger = 'fitness_start,crypto_welcome' WHERE phone = '60123456789';
```

#### Step 2: Automatic Enrollment
The processor automatically:
1. Finds leads with triggers matching sequence entry points
2. Creates sequence_contact entries
3. Sets initial trigger and processing time

#### Step 3: Message Processing
1. Query contacts ready for processing (next_trigger_time <= NOW)
2. Claim batch for specific device (prevents conflicts)
3. Send message via broadcast system
4. Update to next trigger or complete sequence

#### Step 4: Sequence Completion
When `next_trigger` is NULL:
1. Mark sequence_contact as completed
2. Remove trigger from lead
3. Lead can now enter new sequences

## Performance Optimizations

### 1. Database Indexes
```sql
-- Fast trigger lookups
CREATE INDEX idx_leads_trigger ON leads(trigger);
CREATE INDEX idx_seq_contacts_trigger ON sequence_contacts(current_trigger, next_trigger_time);
CREATE INDEX idx_seq_contacts_processing ON sequence_contacts(processing_device_id);
```

### 2. Batch Processing
- Process 1000 contacts per cycle
- 10 parallel workers
- Each device handles ~100 messages

### 3. Device Load Balancing
```go
type DeviceLoad struct {
    MessagesHour      int  // Max ~80-100
    MessagesToday     int  // Max ~800-1000
    CurrentProcessing int  // Max ~50
}
```

### 4. No Complex Queries
- Simple WHERE clauses
- No complex JOINs in hot paths
- Pre-calculated next_trigger_time

## Usage Examples

### 1. Create Sequence with Triggers
```sql
-- Create sequence
INSERT INTO sequences (name, trigger_prefix, is_active) 
VALUES ('30 Day Fitness', 'fitness_', true);

-- Add steps with triggers
INSERT INTO sequence_steps (sequence_id, day_number, trigger, message_text, next_trigger, is_entry_point)
VALUES 
    (seq_id, 1, 'fitness_start', 'Welcome to fitness!', 'fitness_day2', true),
    (seq_id, 2, 'fitness_day2', 'Day 2 workout', 'fitness_day3', false),
    (seq_id, 3, 'fitness_day3', 'Keep going!', NULL, false);
```

### 2. Assign Leads to Sequence
```sql
-- Single sequence
UPDATE leads SET trigger = 'fitness_start' WHERE niche = 'fitness';

-- Multiple sequences
UPDATE leads SET trigger = 'fitness_start,nutrition_start' WHERE id = ?;
```

### 3. Monitor Progress
```sql
-- See sequence progress
SELECT 
    s.name,
    COUNT(*) as total_contacts,
    SUM(CASE WHEN sc.status = 'active' THEN 1 ELSE 0 END) as active,
    SUM(CASE WHEN sc.status = 'completed' THEN 1 ELSE 0 END) as completed
FROM sequences s
JOIN sequence_contacts sc ON sc.sequence_id = s.id
GROUP BY s.name;

-- Check device workload
SELECT 
    device_id,
    messages_hour,
    messages_today,
    is_available
FROM device_load_balance
ORDER BY messages_hour ASC;
```

## Advantages

### 1. Scalability
- Linear scaling with devices
- No bottlenecks at 3000 devices
- Can handle millions of leads

### 2. Simplicity
- One trigger per lead at a time per sequence
- Clear progression path
- Easy to debug

### 3. Flexibility
- Leads can be in multiple sequences
- Easy to pause/resume
- Branching possible with conditional triggers

### 4. Performance
- Minimal database load
- Efficient queries
- Built-in rate limiting

## Monitoring

### Check Processing Status
```sql
-- Active processing
SELECT COUNT(*) FROM sequence_contacts 
WHERE processing_device_id IS NOT NULL;

-- Stuck processing (needs cleanup)
SELECT COUNT(*) FROM sequence_contacts 
WHERE processing_device_id IS NOT NULL 
AND processing_started_at < NOW() - INTERVAL '5 minutes';

-- Retry failures
SELECT contact_phone, retry_count, last_error 
FROM sequence_contacts 
WHERE retry_count > 0 
ORDER BY retry_count DESC;
```

## Best Practices

1. **Set Reasonable Delays**
   - Default: 24 hours between messages
   - Minimum: 1 hour (respect WhatsApp limits)

2. **Monitor Device Health**
   - Check hourly/daily message counts
   - Rotate devices if needed

3. **Clean Triggers**
   - Remove completed triggers from leads
   - Prevent trigger accumulation

4. **Test Sequences**
   - Start with small batches
   - Monitor completion rates
   - Adjust timing as needed

## Troubleshooting

### Messages Not Sending
1. Check if sequence is active
2. Verify trigger exists in sequence_steps
3. Check device availability
4. Look for processing locks

### Slow Processing
1. Check database indexes
2. Verify batch size settings
3. Monitor device workloads
4. Check for stuck processing

### High Retry Rates
1. Check device health
2. Verify phone numbers
3. Look at error messages
4. Consider rate limiting

This trigger-based system provides excellent performance for 3000 devices while maintaining flexibility and ease of use.