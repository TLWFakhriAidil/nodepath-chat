@echo off
echo Committing current changes and pushing...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

REM Add all changes
git add -A

REM Commit with a message about the build error
git commit -m "WIP: Fixing build error - duplicate StopAllWorkers function with unused device variable"

REM Push to trigger the build on Railway
git push origin main --force

echo.
echo ========================================
echo Changes pushed to GitHub!
echo ========================================
echo.
echo The build error is due to:
echo 1. Duplicate StopAllWorkers function
echo 2. Unused 'device' variable in the for loop
echo 3. Orphaned code after function removal
echo.
echo Railway will show the build error and we need to fix it
echo by removing the duplicate function properly.
echo.
pause
