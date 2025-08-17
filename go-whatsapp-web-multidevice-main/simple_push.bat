@echo off
echo Pushing sequence fixes to GitHub...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo Committing changes...
git commit -m "Fix sequence creation: niche, time_schedule, and steps saving"

echo Pushing to main branch...
git push origin main

echo Push complete!
pause
