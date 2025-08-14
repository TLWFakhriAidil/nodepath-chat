@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding files to git...
git add -A

echo Committing changes...
git commit -m "Fix: Disable WebSocket for public view and ensure all API calls use public endpoints filtered by device ID"

echo Pushing to GitHub...
git push origin main

echo Done!
pause