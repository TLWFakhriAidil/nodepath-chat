@echo off
echo ========================================
echo Pushing WhatsApp Session Auto-Load Fix
echo ========================================

echo Adding changes...
git add -A

echo Committing changes...
git commit -m "Fix: Auto-load WhatsApp sessions on server startup

- Load all existing WhatsApp sessions when server starts
- Connect devices that have valid sessions stored
- Update database status based on actual connection state
- No more 'all devices offline' issue after deployment
- Devices with valid sessions will auto-connect without QR scan
- Fixed worker queue size (10000 instead of 1000)"

echo Pushing to GitHub...
git push origin main

echo ========================================
echo Fix pushed successfully!
echo Deploy to Railway to auto-load devices.
echo ========================================
pause
