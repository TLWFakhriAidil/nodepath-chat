@echo off
echo ========================================
echo Reverting to working auto-reconnect version
echo ========================================

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

REM Show current status
echo Current commit:
git log --oneline -1

REM Revert to the working auto-reconnect version
echo.
echo Reverting to commit 4b7d139 (working auto-reconnect)...
git reset --hard 4b7d139

REM Show new status
echo.
echo After revert:
git log --oneline -1

REM Force push to override remote
echo.
echo Force pushing to main branch...
git push --force origin main

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✅ Successfully reverted to working version!
    echo.
    echo Now using the version that:
    echo - Actually attempts to reconnect devices
    echo - Does NOT set devices to offline automatically
    echo - Properly registers with ClientManager
    echo - Sends proper notifications
) else (
    echo.
    echo ❌ Push failed!
)

pause
