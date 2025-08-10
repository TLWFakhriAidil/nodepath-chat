@echo off
echo ========================================
echo Pushing successful build to GitHub
echo ========================================

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

REM Add all changed files
echo Adding changes...
git add src/infrastructure/whatsapp/multidevice_auto_reconnect.go

REM Commit with descriptive message
echo Committing changes...
git commit -m "Fix: Adjust auto-reconnect delay to 30s for proper initialization"

REM Push to main branch
echo Pushing to main branch...
git push origin main

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✅ Successfully pushed to GitHub!
    echo.
    echo Changes:
    echo - Fixed UTF-8 encoding in multidevice_auto_reconnect.go
    echo - Increased initialization delay to 30s for stability
    echo - Build successful with CGO_ENABLED=0
) else (
    echo.
    echo ❌ Push failed! Please check your GitHub credentials.
)

pause
