@echo off
echo Pushing schedule_time fix and optimized worker system...
echo.

cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo.
echo Committing changes...
git commit -m "Fix schedule_time to VARCHAR and optimize worker system for 3000 devices

SCHEDULE_TIME FIXES:
- Changed campaign scheduled_time from TIMESTAMP to VARCHAR(10) for simplicity
- Updated all repository functions to handle string scheduled_time
- Fixed campaign trigger to parse time strings
- Updated REST API to handle string time format
- Added COALESCE for null handling in queries

WORKER SYSTEM OPTIMIZATION:
- Created optimized broadcast manager for 3000+ devices
- Implemented per-device worker with dedicated queue
- Added rate limiting (20/min, 500/hour, 5000/day per device)
- Created worker status tracking and health checks
- Implemented parallel campaign processing
- Added worker repository for monitoring
- Created optimized campaign trigger service

PERFORMANCE FEATURES:
- Max 500 concurrent workers system-wide
- 1000 message queue per worker
- Automatic worker idle timeout
- Round-robin device distribution
- Retry logic with exponential backoff
- Real-time metrics collection
- Database connection pooling

CONFIGURATION:
- Added worker_config.go with all tunable parameters
- Optimized for 200 users x 15 devices each
- Configurable delays and rate limits
- Memory management settings

This makes the system capable of handling massive broadcast campaigns
across thousands of devices simultaneously with proper rate limiting."

echo.
echo Pushing to main branch...
git push origin main

echo.
echo Done! All optimizations have been pushed to GitHub.
echo.
echo NEXT STEPS:
echo 1. Run database migration: fix_schedule_time_and_workers.sql
echo 2. Restart the application
echo 3. Test campaign creation with time values
echo 4. Monitor worker performance in Worker Status tab
pause
