@echo off
echo Pushing latest changes...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

git add -A
git commit -m "Fix schedule_time column mapping in model and repository"
git push origin main

echo Done!
pause
