@echo off
echo ========================================
echo Building Platform-Aware Self-Healing
echo ========================================
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src
set CGO_ENABLED=0

echo Building without CGO...
go build -o whatsapp.exe
if %errorlevel% neq 0 (
    echo Build failed!
    pause
    exit /b 1
)

echo Build successful!
echo.

cd ..

echo Copying executable to root...
copy src\whatsapp.exe whatsapp.exe /Y

echo Adding all changes...
git add -A

echo Committing platform-aware self-healing...
git commit -m "fix: Platform-aware self-healing - only refresh WhatsApp devices

- Platform devices (Wablas/Whacenter) skip refresh and use external API
- Only WhatsApp devices (platform=null) go through self-healing refresh
- Uses exact same refresh logic as working UI refresh button
- Prevents unnecessary refresh attempts for platform devices
- Better error messages distinguishing platform vs WhatsApp devices"

echo Pushing to GitHub...
git push origin main

echo.
echo ========================================
echo ✅ Platform-Aware Self-Healing Deployed!
echo ========================================
echo.
echo Changes:
echo - Platform devices: No refresh, use external API
echo - WhatsApp devices: Self-healing refresh on demand
echo - Same logic as working UI refresh button
echo.
pause
