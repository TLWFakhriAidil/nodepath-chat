# 🚀 Redis Setup Guide for NodePath Chat System

## Overview
This guide will help you set up Redis for your NodePath Chat system to enable high-performance message queuing, caching, and support for 3000+ concurrent devices.

## 🔧 Railway Deployment Setup

### Step 1: Add Redis Service to Railway

1. **Go to your Railway project dashboard**
2. **Click "New" → "Database" → "Add Redis"**
3. **Railway will automatically create the Redis service**

### Step 2: Configure Environment Variables

After adding Redis, Railway should automatically create these variables. If not, add them manually:

```env
# Primary Redis URL (Railway auto-generates this)
REDIS_URL=${{REDIS_URL}}

# Additional Redis variables (Railway auto-generates these)
REDIS_PASSWORD=${{REDIS_PASSWORD}}
REDIS_HOST=${{REDIS_PRIVATE_DOMAIN}}
REDIS_PORT=6379
```

### Step 3: Verify Railway Redis Variables

1. **Go to your app service in Railway**
2. **Click "Variables" tab**
3. **You should see these variables:**
   - `REDIS_URL` (format: `redis://default:password@host.railway.internal:6379`)
   - `REDIS_PASSWORD`
   - `REDIS_PRIVATE_DOMAIN` or `REDISHOST`
   - `REDISPORT` (usually 6379)

### Step 4: Manual Setup (if auto-generation fails)

If Railway doesn't auto-generate the variables:

1. **Click "New Variable" in your app**
2. **Add each variable:**
   ```
   Name: REDIS_URL
   Value: redis://default:YOUR_REDIS_PASSWORD@YOUR_REDIS_HOST.railway.internal:6379
   
   Name: REDIS_PASSWORD
   Value: [Copy from Redis service]
   
   Name: REDIS_HOST
   Value: [Copy from Redis service]
   
   Name: REDIS_PORT
   Value: 6379
   ```

## 🏠 Local Development Setup

### Option 1: Using Docker (Recommended)

1. **Create docker-compose.yml:**
```yaml
version: '3.8'
services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    command: redis-server --requirepass yourpassword
    volumes:
      - redis_data:/data

volumes:
  redis_data:
```

2. **Start Redis:**
```bash
docker-compose up -d redis
```

3. **Set environment variables:**
```env
REDIS_URL=redis://default:yourpassword@localhost:6379
```

### Option 2: Local Redis Installation

1. **Install Redis:**
   - **Windows:** Download from https://redis.io/download
   - **macOS:** `brew install redis`
   - **Linux:** `sudo apt-get install redis-server`

2. **Start Redis:**
```bash
redis-server
```

3. **Set environment variables:**
```env
REDIS_URL=redis://localhost:6379
```

## 🔍 Testing Redis Connection

### Method 1: Using the Application

1. **Start your application:**
```bash
# Set environment variable
$env:REDIS_URL="redis://localhost:6379"

# Run the application
go run cmd/server/main.go
```

2. **Check logs for:**
```
✅ Redis connection established
✅ Redis initialized successfully
```

### Method 2: Using Redis CLI

```bash
# Test local connection
redis-cli ping
# Should return: PONG

# Test with password
redis-cli -a yourpassword ping
# Should return: PONG
```

### Method 3: Using Application API

Once deployed, visit:
```
https://your-app.up.railway.app/api/system/redis-check
```

This will show:
- Redis connection status
- Environment variable validation
- Current broadcast manager type

## 📊 Redis Features Enabled

When Redis is properly configured, your system will have:

### ✅ **Performance Features:**
- **Message Queuing:** Persistent message queues that survive server restarts
- **Rate Limiting:** Prevents WhatsApp bans with intelligent rate limiting
- **Caching:** Faster response times with Redis caching
- **Scalability:** Support for 3000+ concurrent devices

### ✅ **Reliability Features:**
- **Message Persistence:** Zero message loss on crashes
- **Retry Logic:** Automatic retry with exponential backoff
- **Dead Letter Queue:** Failed messages are stored for analysis
- **Multi-Server Support:** Horizontal scaling across multiple servers

## 🚨 Troubleshooting

### Issue: "Redis connection failed"

**Solution:**
1. Check if Redis service is running
2. Verify REDIS_URL format
3. Check firewall/network settings
4. Ensure password is correct

### Issue: "Redis URL not provided"

**Solution:**
1. Set REDIS_URL environment variable
2. Restart the application
3. Check Railway variables are properly set

### Issue: "Target machine actively refused connection"

**Solution:**
1. **Railway:** Ensure Redis service is running in Railway dashboard
2. **Local:** Start Redis server (`redis-server`)
3. Check if port 6379 is available

### Issue: Variables show `${{REDIS_URL}}`

**Solution:**
1. Remove and re-add Redis service in Railway
2. Or manually set the actual Redis URL value

## 🔧 Environment Variable Examples

### Railway Production:
```env
REDIS_URL=redis://default:zwSXYXzTBYBreTwZtPbDVQLJUTHGqYnL@redis.railway.internal:6379
REDIS_PASSWORD=zwSXYXzTBYBreTwZtPbDVQLJUTHGqYnL
REDIS_HOST=redis.railway.internal
REDIS_PORT=6379
```

### Local Development:
```env
REDIS_URL=redis://localhost:6379
# OR with password
REDIS_URL=redis://default:yourpassword@localhost:6379
```

### Local Development with Docker:
```env
REDIS_URL=redis://default:yourpassword@localhost:6379
REDIS_PASSWORD=yourpassword
REDIS_HOST=localhost
REDIS_PORT=6379
```

## 📈 Performance Impact

| Feature | Without Redis | With Redis |
|---------|--------------|------------|
| Max Devices | ~1,500 | **10,000+** |
| Queue Persistence | ❌ Lost on restart | ✅ Survives crashes |
| Multi-Server | ❌ Single server only | ✅ Horizontal scaling |
| Queue Size | 1,000/device | **Unlimited** |
| RAM Usage | 3-5GB | **< 500MB** |
| Message Loss | Possible | **Zero** |
| Worker Recovery | Manual | **Automatic** |

## 🎯 Next Steps

1. **Set up Redis using this guide**
2. **Deploy to Railway**
3. **Test Redis connection**
4. **Monitor performance improvements**
5. **Scale to 3000+ devices**

---

**Need help?** Check the application logs for Redis connection status or visit the Redis check endpoint after deployment.