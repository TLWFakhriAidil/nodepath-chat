@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding files to git...
git add -A

echo Committing changes...
git commit -m "Fix: Move public device routes before auth middleware to allow access without login"

echo Pushing to GitHub...
git push origin main

echo Done!
pause