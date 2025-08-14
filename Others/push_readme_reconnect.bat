@echo off
echo ========================================
echo Pushing README update with reconnection docs
echo ========================================

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

REM Add README
echo Adding README...
git add README.md

REM Commit
echo Committing changes...
git commit -m "Docs: Update README with amazing reconnection feature

- Added detailed explanation of how reconnection works
- Included technical implementation with code examples
- Added database schema showing JID storage
- Highlighted that NO QR SCAN needed if device still linked
- Added to completed features list
- Emphasized the efficiency of direct JID lookup"

REM Push to main branch
echo Pushing to main branch...
git push origin main

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✅ Successfully updated README!
    echo.
    echo The amazing reconnection feature is now documented!
) else (
    echo.
    echo ❌ Push failed!
)

pause
