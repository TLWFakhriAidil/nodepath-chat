@echo off
echo ========================================
echo Pushing online-only reconnect update
echo ========================================

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

REM Add changed file
echo Adding changes...
git add src/infrastructure/whatsapp/multidevice_auto_reconnect.go

REM Commit with descriptive message
echo Committing changes...
git commit -m "Update: Only reconnect devices with 'online' status

- Modified auto-reconnect to only attempt reconnection for devices marked as 'online'
- Skip devices with 'offline' or other statuses
- This ensures only previously active devices are reconnected after restart"

REM Push to main branch
echo Pushing to main branch...
git push origin main

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✅ Successfully pushed to GitHub!
    echo.
    echo Now auto-reconnect will only try to reconnect devices that were online.
) else (
    echo.
    echo ❌ Push failed!
)

pause
