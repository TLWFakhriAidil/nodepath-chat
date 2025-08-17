# Campaigns vs Sequences - Complete Guide

## 🎯 CAMPAIGNS
**Purpose**: One-time bulk broadcast messages to multiple recipients

### Key Characteristics:
- **Single Message**: Sends one message per campaign
- **Scheduled Delivery**: Set a specific date and time
- **Bulk Recipients**: Can send to thousands at once
- **Target Filtering**: Filter by lead status (prospect/customer/all)
- **No Follow-up**: Once sent, campaign is complete

### Database Structure:
```sql
campaigns table:
- id (auto-increment)
- title, message, image_url
- campaign_date, scheduled_at (when to send)
- target_status (who to send to)
- min_delay_seconds, max_delay_seconds (anti-spam)

broadcast_messages table:
- Links to campaign_id
- One record per recipient
- Status: pending → queued → sent/failed
```

### Use Cases:
- Promotional announcements
- Event invitations
- Product launches
- Holiday greetings
- One-time offers

## 🔄 SEQUENCES
**Purpose**: Automated multi-step message flows over days/weeks

### Key Characteristics:
- **Multiple Messages**: Series of messages (Day 1, 3, 7, etc.)
- **Time-Based**: Messages sent based on enrollment date
- **Trigger-Based**: Contacts enrolled when they match conditions
- **Personalized Journey**: Each contact progresses individually
- **Status Progression**: Leads move from COLD → WARM → HOT

### Database Structure:
```sql
sequences table:
- id (UUID)
- name, description
- trigger (enrollment condition)
- total_days (sequence duration)

sequence_steps table:
- sequence_id (parent)
- day_number (when to send: 1, 3, 7, etc.)
- content, media_url
- trigger conditions for each step

sequence_contacts table:
- Tracks each enrolled contact
- current_step (progress)
- next_trigger_time
- status (active/completed/failed)

broadcast_messages table:
- Links to sequence_id AND sequence_stepid
- Same status flow as campaigns
```

### Use Cases:
- Lead nurturing
- Onboarding flows
- Educational series
- Follow-up sequences
- Relationship building

## 📊 Key Differences

| Feature | Campaign | Sequence |
|---------|----------|----------|
| Messages | Single | Multiple (series) |
| Timing | All at once | Spread over days/weeks |
| Personalization | Same for all | Individual progress |
| Enrollment | Manual selection | Automatic (trigger-based) |
| Tracking | Simple (sent/failed) | Complex (step progress) |
| Purpose | Broadcast | Nurture |
| Flexibility | Low | High |

## 🔧 Technical Implementation

### Campaigns:
1. Create campaign with message
2. System creates broadcast_messages for all matching leads
3. Worker pool processes at scheduled time
4. Anti-spam delays between sends
5. Complete when all sent

### Sequences:
1. Create sequence with multiple steps
2. Leads enrolled when trigger matches
3. First message scheduled immediately
4. After each send, calculate next step time
5. Continue until sequence completes
6. Each contact progresses independently

Both use the same `broadcast_messages` table and worker pool system for actual sending!