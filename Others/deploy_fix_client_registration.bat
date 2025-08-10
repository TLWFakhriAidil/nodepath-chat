@echo off
echo ========================================
echo Fixing WhatsApp Client Registration Issue
echo ========================================

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo Creating commit...
git commit -m "Fix: WhatsApp client registration and worker creation

- Added detailed debugging for client registration issues
- Added GetClientCount and GetAllClients methods to ClientManager
- Enhanced logging to show all registered clients when worker creation fails
- Fixed queue checking to run every 100ms for 3000 device support
- QR event spam already fixed (only logs non-empty events)
- Better error messages to diagnose why workers aren't being created"

echo Pushing to GitHub...
git push origin main

echo.
echo ========================================
echo Push Complete!
echo ========================================
echo.
echo This will help debug why workers aren't being created:
echo 1. Shows total registered clients
echo 2. Lists all registered device IDs
echo 3. Helps identify if client registration is failing
echo 4. Queue checking now runs every 100ms
echo.
echo After deployment, check logs for:
echo - "Total registered clients in ClientManager: X"
echo - "Registered device ID: xxx"
echo.
pause
