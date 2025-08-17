@echo off
echo ========================================
echo Pushing Critical Duplicate Registration Fix
echo ========================================
echo.

REM Add all changes
git add -A

REM Commit with descriptive message
git commit -m "fix: Prevent duplicate device registration and disconnection

- Fixed duplicate keepalive starts by checking if already active
- Prevented client replacement when same client is added again
- Register with DeviceManager BEFORE event handlers to ensure proper flow
- Fixed 'connection unregistered' issue after 1 minute

The issue was that device was being registered twice:
1. Once during reconnect
2. Again when Connected event fired
This caused conflicts and premature disconnection.

Now devices will stay connected properly for 24/7 operation."

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
echo Successfully pushed duplicate registration fix!
echo ========================================
echo.
echo This fix ensures:
echo - No duplicate registration
echo - Devices stay connected 24/7
echo - No more "device not found" after 1 minute
echo.
pause
