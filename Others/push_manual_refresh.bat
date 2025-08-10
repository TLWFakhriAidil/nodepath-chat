@echo off
echo ========================================
echo Pushing manual refresh feature
echo ========================================

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

REM Add all changed files
echo Adding changes...
git add src/cmd/rest.go
git add src/views/dashboard.html
git add README.md

REM Commit
echo Committing changes...
git commit -m "Feature: Manual device refresh instead of auto-reconnect

- Disabled automatic reconnection on server startup
- Added 'Refresh' button in device dropdown menu
- Uses existing /api/devices/:deviceId/connect endpoint
- Attempts to reconnect using stored WhatsApp session
- Shows clear feedback if session expired (prompts for QR scan)
- Gives users full control over when to reconnect devices
- Updated README with manual refresh documentation"

REM Push to main branch
echo Pushing to main branch...
git push origin main

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✅ Successfully pushed manual refresh feature!
    echo.
    echo Changes:
    echo - Auto-reconnect is now DISABLED
    echo - Users can manually refresh devices using the new button
    echo - Better control over device connections
) else (
    echo.
    echo ❌ Push failed!
)

pause
