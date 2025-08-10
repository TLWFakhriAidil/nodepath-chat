@echo off
echo Pushing fixes for schedule_time display and View button...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo Committing changes...
git commit -m "Fix schedule_time display and improve error logging

- Fixed schedule_time display by checking both lowercase and camelCase
- Added detailed error logging for GetSequenceByID
- This should show schedule_time in UI and help debug View button 404"

echo Pushing to main branch...
git push origin main

echo Push complete!
pause
