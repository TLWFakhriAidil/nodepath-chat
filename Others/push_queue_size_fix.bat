@echo off
echo ========================================
echo Pushing Worker Queue Size Fix
echo ========================================

echo Adding changes...
git add -A

echo Committing changes...
git commit -m "Fix worker queue size: Use config value 10000 instead of hardcoded 1000

- Updated ultra_scale_broadcast_manager.go to use config.WorkerQueueSize
- Updated device_worker.go to use config.WorkerQueueSize  
- This should fix 'timeout queueing message to worker' errors
- Queue size increased from 1000 to 10000 as per 5K optimization"

echo Pushing to GitHub...
git push origin main

echo ========================================
echo Fix pushed successfully!
echo Deploy to Railway to apply the changes.
echo ========================================
pause
