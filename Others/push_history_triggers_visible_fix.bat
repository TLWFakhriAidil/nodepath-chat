@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo Committing changes...
git commit -m "Fix: History Triggers filter now visible in Lead Management" -m "- Updated buildTriggerFilters function to include History Triggers option" -m "- History Triggers filter is now properly displayed after page load" -m "- Click handler properly switches between normal and history mode"

echo Pushing to GitHub...
git push origin main

echo Done!
pause
