@echo off
echo Checking git status and pushing if needed...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Current status:
git status

echo.
echo Adding any modified files...
git add -A

echo.
echo Committing if there are changes...
git commit -m "Fix schedule_time column mapping in model and repository"

echo.
echo Pushing to main branch...
git push origin main

echo.
echo Done!
pause
