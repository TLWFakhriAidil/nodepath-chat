@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding README changes...
git add README.md

echo Committing changes...
git commit -m "Update README with August 3 critical fixes summary" -m "- Added critical message fix where GetPendingMessages wasn't appending messages" -m "- Added anti-spam flow fix for double application" -m "- Added platform device anti-spam support"

echo Pushing to GitHub...
git push origin main

echo Done!
pause
