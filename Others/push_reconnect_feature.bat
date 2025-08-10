@echo off
echo ========================================
echo Pushing device reconnection implementation
echo ========================================

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

REM Add all changed files
echo Adding changes...
git add src/ui/rest/device_reconnect.go
git add src/ui/rest/app.go
git add src/views/dashboard.html

REM Commit
echo Committing changes...
git commit -m "Feature: Proper device reconnection without QR scan

- Created ReconnectDeviceSession endpoint that actually tries to reconnect
- Searches PostgreSQL store for existing WhatsApp sessions by phone number
- If session found, creates client and attempts to connect
- Only asks for QR scan if no session found or connection fails
- Dashboard now shows specific messages for different scenarios
- Should reconnect devices that are still linked in WhatsApp"

REM Push to main branch
echo Pushing to main branch...
git push origin main

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✅ Successfully pushed reconnection feature!
    echo.
    echo Now devices can reconnect without QR scan if:
    echo - They are still linked in WhatsApp
    echo - The session data exists in PostgreSQL
) else (
    echo.
    echo ❌ Push failed!
)

pause
