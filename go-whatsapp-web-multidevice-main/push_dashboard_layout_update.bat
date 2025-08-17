@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding files to git...
git add -A

echo Committing changes...
git commit -m "Update: Public device view now uses exact same layout as main dashboard"

echo Pushing to GitHub...
git push origin main

echo Done!
pause