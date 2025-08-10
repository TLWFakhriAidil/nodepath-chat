@echo off
echo ==========================================
echo DEPLOYING REDIS-OPTIMIZED WHATSAPP SYSTEM
echo For 3000+ Device Support
echo ==========================================
echo.

REM Navigate to project directory
cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo [1/5] Adding all changes...
git add .

echo.
echo [2/5] Committing changes...
git commit -m "feat: Ultra Scale Redis implementation for 3000+ devices

- Implemented UltraScaleRedisManager with optimizations for 3000 devices
- Device-specific queues for better distribution
- Worker pooling and efficient resource management
- Batched metrics for reduced Redis load
- Health monitoring with automatic recovery
- Distributed lock system for multi-server support
- Maximum 3000 concurrent workers
- Optimized Redis connection pooling (100 connections)
- Automatic worker lifecycle management
- Dead letter queue per device
- Exponential backoff for failed messages"

echo.
echo [3/5] Pushing to remote repository...
git push origin main

echo.
echo [4/5] Deployment Instructions:
echo ==========================================
echo Your Redis-optimized system is ready!
echo.
echo IMPORTANT: Redis will be automatically detected when deployed to Railway
echo with the following environment variables set:
echo.
echo REDIS_URL=redis://default:zwSXYXzTBYBreTwZtPbDVQLJUTHGqYnL@redis.railway.internal:6379
echo.
echo The system will:
echo - Support 3000+ devices simultaneously
echo - Use device-specific queues
echo - Distribute load across multiple servers
echo - Persist messages in Redis
echo - Auto-recover from crashes
echo.

echo [5/5] Next Steps:
echo ==========================================
echo 1. Go to Railway dashboard
echo 2. Check deployment logs
echo 3. Look for: "Successfully connected to Redis (Ultra Scale Mode)"
echo 4. Visit /api/system/redis-check to verify
echo.

echo ==========================================
echo DEPLOYMENT COMPLETE!
echo ==========================================
pause
