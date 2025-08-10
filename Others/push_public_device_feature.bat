@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding files to git...
git add -A

echo Committing changes...
git commit -m "Feature: Add public device view - accessible via /device/{deviceId} URL without authentication"

echo Pushing to GitHub...
git push origin main

echo Done!
pause