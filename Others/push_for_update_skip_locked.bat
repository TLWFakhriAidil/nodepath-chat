@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding files to git...
git add -A

echo Committing changes...
git commit -m "Fix: Implement FOR UPDATE SKIP LOCKED in GetPendingMessagesAndLock to prevent duplicate messages"

echo Pushing to GitHub...
git push origin main

echo Done!
pause