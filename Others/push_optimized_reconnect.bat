@echo off
echo ========================================
echo Pushing optimized reconnection logic
echo ========================================

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

REM Add changed file
echo Adding changes...
git add src/ui/rest/device_reconnect.go

REM Commit
echo Committing changes...
git commit -m "Optimize: Use device JID to query WhatsApp session directly

- Now queries whatsmeow_sessions table using the exact JID from user_devices
- No longer searches through all devices - direct lookup by JID
- Uses container.GetDevice(ctx, jid) to get specific device by JID
- Fixed field name from device.Jid to device.JID
- More efficient and accurate session restoration"

REM Push to main branch
echo Pushing to main branch...
git push origin main

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✅ Successfully pushed optimized reconnection!
    echo.
    echo Now reconnection:
    echo - Uses the exact JID stored in user_devices table
    echo - Queries whatsmeow_sessions directly by JID
    echo - Much faster and more accurate
) else (
    echo.
    echo ❌ Push failed!
)

pause
