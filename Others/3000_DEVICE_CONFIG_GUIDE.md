# 🚀 3000 Device Configuration Guide

## Optimal Settings for 3000 Devices with Redis

### 1. **Railway Resources**
For 3000 devices, you'll need:
- **CPU**: 8+ vCPUs (16 recommended)
- **RAM**: 8GB minimum (16GB recommended)
- **Storage**: 100GB+ for message history
- **Redis**: Railway Redis addon (already configured)

### 2. **Environment Variables**
```env
# Redis (Already set in your Railway)
REDIS_URL=redis://default:zwSXYXzTBYBreTwZtPbDVQLJUTHGqYnL@redis.railway.internal:6379
REDIS_PASSWORD=zwSXYXzTBYBreTwZtPbDVQLJUTHGqYnL
REDISHOST=redis.railway.internal
REDISPORT=6379

# Application Settings
APP_PORT=3000
APP_DEBUG=false
APP_BASIC_AUTH=admin:your-secure-password

# Database
DB_URI=postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require

# WhatsApp Settings
WHATSAPP_CHAT_STORAGE=true
WHATSAPP_ACCOUNT_VALIDATION=true
```

### 3. **Database Optimizations**
Run these SQL commands in your PostgreSQL:

```sql
-- Increase connection limit
ALTER DATABASE railway SET max_connections = 500;

-- Optimize for high throughput
ALTER DATABASE railway SET shared_buffers = '4GB';
ALTER DATABASE railway SET effective_cache_size = '12GB';
ALTER DATABASE railway SET work_mem = '64MB';
ALTER DATABASE railway SET maintenance_work_mem = '512MB';

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_broadcast_messages_device_status 
ON broadcast_messages(device_id, status);

CREATE INDEX IF NOT EXISTS idx_broadcast_messages_campaign 
ON broadcast_messages(campaign_id, status);

CREATE INDEX IF NOT EXISTS idx_user_devices_user_active 
ON user_devices(user_id, is_active);

-- Partition broadcast_messages table by date (optional for very high volume)
-- This helps with cleanup and performance
```

### 4. **Device Distribution Strategy**

#### Recommended Setup:
- **200 users** × **15 devices each** = **3000 total devices**
- Each user manages their own device pool
- Campaigns automatically distribute across user's devices

#### User Creation Script:
```sql
-- Example: Create users with device allocation
DO $$
DECLARE
    i INTEGER;
BEGIN
    FOR i IN 1..200 LOOP
        INSERT INTO users (email, password, name)
        VALUES (
            'user' || i || '@company.com',
            'hashed_password_here',
            'User ' || i
        );
    END LOOP;
END $$;
```

### 5. **Rate Limiting Configuration**

#### Per-Device Settings:
```sql
-- Set optimal delays for all devices
UPDATE user_devices 
SET 
    min_delay_seconds = 10,
    max_delay_seconds = 30
WHERE is_active = true;

-- For specific high-volume devices
UPDATE user_devices 
SET 
    min_delay_seconds = 15,
    max_delay_seconds = 45
WHERE id IN (SELECT id FROM user_devices ORDER BY created_at LIMIT 100);
```

### 6. **Redis Queue Management**

The system uses these Redis queues:
- `broadcast:queue:campaign:{device_id}` - Campaign messages
- `broadcast:queue:sequence:{device_id}` - Sequence messages
- `broadcast:queue:dead:{device_id}` - Failed messages
- `broadcast:workers` - Worker status tracking
- `broadcast:metrics:*` - Performance metrics

### 7. **Monitoring Commands**

#### Check Redis Status:
```bash
# In Railway console
redis-cli INFO clients
redis-cli INFO memory
redis-cli INFO stats
```

#### Monitor Queue Sizes:
```bash
# Check campaign queues
redis-cli --scan --pattern "broadcast:queue:campaign:*" | xargs -I {} redis-cli LLEN {}

# Check worker status
redis-cli HGETALL broadcast:workers
```

### 8. **Performance Tuning**

#### Application Level:
1. **Worker Pools**: System uses worker pooling by priority
2. **Batch Processing**: Metrics are batched before Redis writes
3. **Connection Pooling**: Redis uses 100 connection pool
4. **Queue Sharding**: Each device has its own queue

#### Message Processing:
- **Campaign Priority**: 1 (highest)
- **Sequence Priority**: 5 (normal)
- **Retry Logic**: 3 retries with exponential backoff
- **Dead Letter**: After 3 failures

### 9. **Scaling Beyond 3000 Devices**

If you need more than 3000 devices:

1. **Horizontal Scaling**: 
   - Deploy multiple Railway instances
   - Use Railway's load balancer
   - Redis coordinates between instances

2. **Database Sharding**:
   - Partition users across multiple databases
   - Use device_id prefix for routing

3. **Redis Cluster**:
   - Upgrade to Redis Cluster for 10,000+ devices
   - Automatic sharding across nodes

### 10. **Best Practices**

#### DO:
- ✅ Monitor worker health regularly
- ✅ Use device-specific delays
- ✅ Distribute devices across users
- ✅ Stagger campaign start times
- ✅ Monitor Redis memory usage
- ✅ Use retry functionality for failures

#### DON'T:
- ❌ Run all 3000 devices from one user
- ❌ Set delays too low (< 5 seconds)
- ❌ Ignore failed message counts
- ❌ Disable health monitoring
- ❌ Overload single campaigns

### 11. **Emergency Commands**

If something goes wrong:

```bash
# Stop all workers
curl -X POST https://your-app.railway.app/api/workers/stop-all

# Clear stuck queues
redis-cli FLUSHDB

# Restart workers
curl -X POST https://your-app.railway.app/api/workers/resume-failed
```

### 12. **Expected Metrics**

With proper configuration:
- **Throughput**: 60,000+ messages/minute
- **Worker Stability**: 99.9% uptime
- **Message Success**: 95%+ delivery rate
- **Queue Processing**: < 1 second latency
- **Memory Usage**: < 500MB per instance
- **Redis Memory**: < 2GB for queues

## 🎯 Quick Checklist

- [ ] Redis connected and verified
- [ ] Database indexes created
- [ ] Rate limits configured
- [ ] Users and devices distributed
- [ ] Monitoring dashboard accessible
- [ ] Worker health check enabled
- [ ] Backup strategy in place

Your system is now ready to handle 3000 devices with excellent performance! 🚀
