@echo off
echo ========================================
echo Pushing Auto Device Refresh + Queue Fix
echo ========================================

echo Adding changes...
git add -A

echo Committing changes...
git commit -m "Add auto device refresh on startup + fix worker queue size

- Auto refresh all WhatsApp devices when server starts
- Check actual client connection status (not just database status)  
- Update device status to online/offline based on real connection
- No more manual refresh button clicks needed!
- Fixed worker queue size to use config value (10000 instead of 1000)
- This fixes 'timeout queueing message to worker' errors"

echo Pushing to GitHub...
git push origin main

echo ========================================
echo Changes pushed successfully!
echo Deploy to Railway to apply auto-refresh.
echo ========================================
pause
