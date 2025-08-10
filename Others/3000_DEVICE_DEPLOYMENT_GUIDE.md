# WhatsApp 3000 Device Configuration Guide

## Summary of Changes Made

### 1. **Device Health Monitor** (`device_health_monitor.go`)
- ✅ Increased monitor interval: 30s → 2 minutes
- ✅ Increased failure threshold: 3 → 30 checks (15 minutes)
- ✅ Increased removal timeout: 12 hours → 24 hours
- ✅ Delayed initial check: 10s → 30s

### 2. **Keepalive Manager** (`keepalive_manager.go`)
- ✅ Increased min interval: 45s → 3 minutes
- ✅ Increased max interval: 90s → 5 minutes
- ✅ Increased activity threshold: 3 min → 10 minutes

### 3. **Client Manager** (`client_manager.go`)
- ✅ Added connection limit check (1500 max concurrent)
- ✅ Added rejection logging when limit reached

### 4. **New Components Added**
- ✅ `device_optimizer.go` - Connection optimization logic
- ✅ `redis_queue_manager.go` - Redis-based message queuing

## Environment Setup for 3000 Devices

```bash
# Required Environment Variables
export DB_URI="postgresql://user:pass@localhost:5432/whatsapp_3000"
export REDIS_URL="redis://localhost:6379/0"
export APP_PORT="3000"
export WHATSAPP_LOG_LEVEL="INFO"  # Not DEBUG

# Performance Tuning (Optional)
export GOMAXPROCS=8                # Use 8 CPU cores
export DEVICE_BATCH_SIZE=100       # Process devices in batches
export MESSAGE_RATE_LIMIT=20       # Messages per minute per device
```

## Database Optimizations

```sql
-- Run these on your PostgreSQL database
CREATE INDEX idx_user_devices_status ON user_devices(status);
CREATE INDEX idx_user_devices_user_status ON user_devices(user_id, status);
CREATE INDEX idx_broadcast_messages_status_device ON broadcast_messages(status, device_id);
CREATE INDEX idx_broadcast_messages_scheduled ON broadcast_messages(scheduled_at) WHERE status = 'pending';

-- Increase connection pool
ALTER SYSTEM SET max_connections = 500;
ALTER SYSTEM SET shared_buffers = '4GB';
ALTER SYSTEM SET effective_cache_size = '12GB';
```

## Redis Configuration

```conf
# /etc/redis/redis.conf
maxclients 10000
timeout 300
tcp-keepalive 60
maxmemory 4gb
maxmemory-policy allkeys-lru
save ""  # Disable persistence for performance
```

## Deployment Steps

### 1. Build the Application
```bash
cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
build_local.bat
```

### 2. Start Services
```bash
# Start PostgreSQL
pg_ctl start -D "C:\Program Files\PostgreSQL\14\data"

# Start Redis
redis-server

# Start WhatsApp
whatsapp.exe rest
```

### 3. Monitor Performance
- Watch logs for connection limit warnings
- Monitor Redis queue sizes
- Check database connection count
- Track message delivery rates

## Scaling Strategy

### Phase 1: 0-500 Devices
- Start with default settings
- Monitor system resources
- Ensure stable operation

### Phase 2: 500-1000 Devices
- Enable Redis queuing
- Increase database connections
- Monitor message delays

### Phase 3: 1000-2000 Devices
- Enable connection optimizer
- Implement device batching
- Add more workers if needed

### Phase 4: 2000-3000 Devices
- Maximum optimization mode
- Consider horizontal scaling
- Monitor closely for issues

## Troubleshooting

### High CPU Usage
- Reduce health check frequency
- Increase batch sizes
- Enable Redis queuing

### Database Connection Errors
- Increase PostgreSQL max_connections
- Use connection pooling
- Check for connection leaks

### Message Delays
- Check Redis queue sizes
- Increase worker count
- Verify rate limiting

### Frequent Disconnections
- Check keepalive logs
- Verify network stability
- Monitor WhatsApp rate limits

## Best Practices

1. **Gradual Scaling**
   - Add devices in batches of 100
   - Wait for stability before adding more
   - Monitor resource usage

2. **Resource Monitoring**
   - CPU usage < 80%
   - Memory usage < 80%
   - Database connections < 80% of max

3. **Rate Limiting**
   - 20 messages/minute per device
   - 5-minute cooldown between reconnects
   - Respect WhatsApp's limits

4. **Backup Strategy**
   - Regular database backups
   - Export device sessions
   - Document configuration

## Performance Metrics

With these optimizations, you should achieve:
- **Connection Stability**: 95%+ uptime
- **Message Throughput**: 20,000-40,000 messages/hour
- **Resource Usage**: Moderate (8 cores, 16GB RAM)
- **Database Load**: Manageable with proper indexing

## Next Steps

1. Test with 50 devices first
2. Monitor for 24 hours
3. Gradually increase to 500
4. Implement monitoring dashboard
5. Scale to target 3000

Remember: WhatsApp has rate limits and anti-spam measures. Always test thoroughly and scale gradually.
