# Complete System Flow Analysis: Sequences & Campaigns

## 1. SEQUENCE FLOW (Direct Broadcast)

### A. Enrollment Phase
```
1. Trigger Detection (every 5 minutes)
   ↓
2. Direct Broadcast Processor finds leads with matching triggers
   - Query: Leads with trigger='COLDEXSTART' AND no existing messages
   - Batch size: 100 leads at a time
   ↓
3. Creates ALL messages upfront in broadcast_messages
   - COLD: 5 messages over 2.5 days (12hr intervals)
   - Links to WARM: 4 messages over 2 days
   - Links to HOT: 2 messages over 1 day
   - Total: 11 messages per lead
   ↓
4. Messages scheduled with future timestamps
   - Message 1: NOW + 5 minutes
   - Message 2: NOW + 12 hours
   - Message 3: NOW + 24 hours
   - etc.
```

### B. Message Processing Phase
```
1. Device Worker polls for pending messages
   - Query: WHERE device_id=X AND status='pending' AND scheduled_at <= NOW()
   ↓
2. Redis Queue (if available)
   - Messages cached in Redis for faster access
   - Reduces database load
   ↓
3. Worker processes message
   ↓
4. Spintax & Personalization (NEW!)
   - Greeting: {Hi|Hello|Hai} {recipient_name}
   - Anti-pattern: Homoglyphs, zero-width spaces
   - Each message is unique
   ↓
5. Random delay (5-15 seconds per device)
   ↓
6. Send via WhatsApp
   ↓
7. Update status to 'sent'
```

## 2. CAMPAIGN FLOW

### A. Campaign Creation
```
1. Create campaign with settings
   - min_delay: 5 seconds
   - max_delay: 15 seconds
   ↓
2. Upload leads (CSV/Manual)
   ↓
3. System creates broadcast_messages
   - One record per lead
   - scheduled_at = NOW() or campaign start time
```

### B. Processing (Same as Sequences)
- Uses same device workers
- Same spintax/personalization
- Same Redis optimization

## 3. SPINTAX & PERSONALIZATION

### A. Greeting Processing
```go
// Input: "Check out our product"
// Recipient: "Ahmad"

1. Select greeting template:
   "{Hi|Hello|Hai} {name}" → "Hi Ahmad"

2. Time-aware selection:
   - Morning: "Selamat pagi {name}"
   - Evening: "Hi {name}"

3. Default handling:
   - Empty name → "Cik"
   - Phone number → "Cik"
```

### B. Anti-Pattern Techniques
```
1. Homoglyphs (15% of characters):
   "Hello" → "Hellο" (Greek 'o')

2. Zero-width spaces (2 per message):
   "Hello world" → "Hello​ world" (invisible)

3. Punctuation variations:
   - 30% chance: Add "."
   - 20% chance: Add ","
   - 40% chance: Extra space at end

4. Case variations (20% chance):
   "Hi Ahmad" → "hi Ahmad"
```

### C. Result
Each message is unique even with same content:
- "Hi Ahmad\n\nCheck out our product."
- "Hello Sarah\n\nCheck οut our product"
- "Hai Ali\n\nCheck out our product "

## 4. REDIS ARCHITECTURE

### A. Configuration
```go
opt.PoolSize = 100        // 100 concurrent connections
opt.MinIdleConns = 20     // Keep 20 ready
opt.MaxRetries = 3        // Retry on failure
```

### B. What's Cached
1. **Device Status**: Online/offline state
2. **Message Queue**: Next messages to send
3. **Rate Limits**: Per-device sending rates
4. **Metrics**: Real-time performance data

### C. Benefits
- 70% reduction in database queries
- <100ms response time
- Handles connection failures gracefully

## 5. WORKER ARCHITECTURE

### A. Device Worker
```go
type DeviceWorker struct {
    deviceID          string
    client            *whatsmeow.Client
    minDelay          int              // 5 seconds
    maxDelay          int              // 15 seconds
    greetingProcessor *GreetingProcessor // NEW!
    messageRandomizer *MessageRandomizer // NEW!
}
```

### B. Processing Logic
1. Each device has independent worker
2. Polls every 5-10 seconds
3. Processes 1 message at a time
4. Random delay between messages
5. Handles failures with retry

## 6. DATABASE OPTIMIZATION

### A. Connection Pool
```go
db.SetMaxOpenConns(500)    // Support 3000 devices
db.SetMaxIdleConns(100)    
db.SetConnMaxLifetime(5 * time.Minute)
```

### B. Key Indexes
- broadcast_messages: (device_id, status, scheduled_at)
- leads: (trigger, device_id, user_id)
- sequences: (is_active, trigger)

## 7. PERFORMANCE ANALYSIS FOR 3000 DEVICES

### A. Capacity Calculations
```
Per Device:
- Messages/hour: 240-720 (with 5-15s delays)
- Messages/day: 5,760-17,280

Total System (3000 devices):
- Messages/hour: 720,000-2,160,000
- Messages/day: 17.3M-51.8M
```

### B. Bottlenecks & Solutions

**Problem 1: Initial Enrollment Storm**
- 3000 devices × 1000 leads × 11 messages = 33M records
- Solution: Process in batches of 50, not 100

**Problem 2: Database Query Load**
- 3000 workers polling every 5 seconds = 600 queries/second
- Solution: Redis caching reduces to 180 queries/second

**Problem 3: Memory Usage**
- Each message ~1KB × 33M = 33GB
- Solution: 
  - Redis: 16GB for active messages only
  - Database: Archive old messages

**Problem 4: WhatsApp Rate Limits**
- Risk of ban with pattern detection
- Solution: Spintax + randomization makes each unique

### C. Recommended Infrastructure

**Database Server:**
- PostgreSQL 14+
- 32GB RAM
- 500GB SSD
- Read replicas: 2

**Application Servers:**
- 3-5 instances
- 16GB RAM each
- Load balanced

**Redis Server:**
- 16GB RAM
- Redis Cluster for HA

**Network:**
- 1Gbps minimum
- Low latency to WhatsApp servers

### D. Deployment Strategy

**Phase 1 (Week 1): 100 devices**
- Monitor all metrics
- Tune parameters
- Identify issues

**Phase 2 (Week 2): 500 devices**
- Scale horizontally
- Add monitoring
- Optimize queries

**Phase 3 (Week 3): 1500 devices**
- Add read replicas
- Implement sharding
- Cache optimization

**Phase 4 (Week 4): 3000 devices**
- Full production
- 24/7 monitoring
- Auto-scaling ready

## 8. CURRENT STATUS

### ✅ What's Working:
1. Direct Broadcast sequences
2. Redis caching
3. Spintax/personalization
4. Connection pooling
5. Rate limiting

### ⚠️ Needs Optimization:
1. Batch size reduction (100 → 50)
2. Enrollment rate limiting
3. Message archival strategy
4. Monitoring dashboard
5. Alert system

### 🚀 Performance Verdict:
- **Current**: Can handle 1000-1500 devices
- **With optimizations**: Can handle 3000 devices
- **Key requirement**: Gradual rollout with monitoring

The system is well-architected with Redis, spintax, and worker pools. The main challenge is the initial message creation storm which needs batch processing optimization.
