@echo off
cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo Committing changes...
git commit -m "Fix AI Campaign migration - use UUID types to match users and devices tables"

echo Pushing to main branch...
git push origin main

echo Done!
echo.
echo NOTE: The AI Campaign tables will now be created correctly with proper foreign keys.
echo.
pause