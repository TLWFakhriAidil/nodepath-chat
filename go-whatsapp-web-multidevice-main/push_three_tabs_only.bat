@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding files to git...
git add -A

echo Committing changes...
git commit -m "Update: Public device view now shows only 3 tabs (Devices, Campaign Summary, Sequence Summary) with data filtered by device ID from URL"

echo Pushing to GitHub...
git push origin main

echo Done!
pause