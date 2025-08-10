@echo off
echo ========================================
echo Reverting to last working commit
echo ========================================

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

REM Revert to the commit before auto-reconnect changes
echo Reverting to commit 7fd11bf...
git reset --hard 7fd11bf

REM Force push to override remote
echo Force pushing to main branch...
git push --force origin main

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✅ Successfully reverted to last stable version!
    echo.
    echo Reverted all auto-reconnect changes.
) else (
    echo.
    echo ❌ Revert failed!
)

pause
