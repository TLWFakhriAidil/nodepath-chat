@echo off
echo ========================================
echo Pushing Worker Optimization Fixes
echo ========================================

echo Adding changes...
git add -A

echo Committing changes...
git commit -m "Optimize worker management to prevent timeouts

- Implemented multiple workers per device (5 workers as configured)
- Increased queue timeout from 5s to 30s (configurable)
- Added load balancer for better message distribution
- Fixed worker pool to use device groups instead of single workers
- Queue size already using config value (10,000)

This should significantly reduce 'timeout queueing message to worker' errors
especially when handling large batches like 678 sequence messages."

echo Pushing to GitHub...
git push origin main

echo ========================================
echo Worker optimization pushed!
echo Deploy to Railway to apply the fixes.
echo ========================================
pause
