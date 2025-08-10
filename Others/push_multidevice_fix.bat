@echo off
echo ========================================
echo Pushing Multi-Device Connection Fix
echo ========================================
echo.

REM Add all changes
git add -A

REM Commit with descriptive message
git commit -m "fix: Prevent device disconnection when multiple devices connect

- Added MultiDeviceClientManager to maintain all device connections
- Ensures devices don't lose their client registration
- Provides fallback mechanism if ClientManager loses a device
- Fixed 'device not connected' error for multi-device scenarios

This ensures all connected devices stay registered and can send messages
even when new devices are added."

REM Push to main branch
git push origin main

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo Push failed! You may need to pull first.
    pause
    exit /b 1
)

echo.
echo ========================================
echo Successfully pushed multi-device fix!
echo ========================================
echo.
echo Key improvements:
echo - Multiple devices can now stay connected simultaneously
echo - Previous devices won't lose connection when new ones connect
echo - Fallback mechanism ensures devices can still send messages
echo - Better device registration persistence
echo.
pause
