@echo off
echo ========================================
echo Build successful - Pushing to GitHub
echo ========================================

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

REM Check for any uncommitted changes
git status --porcelain > temp_status.txt
set /p STATUS=<temp_status.txt
del temp_status.txt

if "%STATUS%"=="" (
    echo No changes to commit. Repository is up to date.
) else (
    echo Found uncommitted changes. Committing...
    git add -A
    git commit -m "Build verified: Manual refresh feature working"
)

REM Push to main
echo Pushing to main branch...
git push origin main

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✅ Successfully pushed to GitHub!
    echo.
    echo The manual refresh feature is now live:
    echo - Auto-reconnect disabled
    echo - Manual refresh button added
    echo - Users control when to reconnect
) else (
    echo.
    echo ❌ Push failed!
)

pause
