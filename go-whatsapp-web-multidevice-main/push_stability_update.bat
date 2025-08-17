@echo off
echo ========================================
echo Committing and Pushing Stability Updates
echo ========================================
echo.

REM Add all changes
git add -A

REM Commit with descriptive message
git commit -m "feat: Add smart keepalive mechanism for 3000 device stability

- Implemented KeepaliveManager for null platform devices only
- Random 45-90 second intervals to mimic WhatsApp Web behavior
- Activity tracking prevents keepalive during active messaging
- Extended grace period to 3 minutes before marking devices offline
- Auto-reconnect always enabled for all devices
- Improved connection stability for daily broadcasting

This update ensures devices stay connected for extended periods,
supporting 3000+ devices for daily broadcast operations without
disconnections (except for bans)."

REM Push to main branch
git push origin main

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo Push failed! You may need to pull first:
    echo git pull origin main
    echo.
    pause
    exit /b 1
)

echo.
echo ========================================
echo Successfully pushed to GitHub!
echo ========================================
echo.
pause
