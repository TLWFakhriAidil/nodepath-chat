@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding files to git...
git add -A

echo Committing changes...
git commit -m "Update: Public device view now uses exact duplicate of dashboard.html with only 3 tabs"

echo Pushing to GitHub...
git push origin main

echo Done!
pause