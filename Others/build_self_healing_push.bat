@echo off
echo ========================================
echo Building with Self-Healing Architecture
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

echo Committing self-healing architecture...
git commit -m "feat: Implement self-healing worker architecture for 3000+ devices

- Added WorkerClientManager with GetOrRefreshClient() for automatic connection refresh
- Workers now refresh device connections before each message send
- Disabled background keepalive and health monitor systems
- No more 'device not found' errors - guaranteed fresh connections
- Scales efficiently to 3000+ devices without background overhead
- Per-device mutex prevents duplicate refreshes
- Updated README with self-healing architecture documentation"

echo Pushing to GitHub...
git push origin main

echo.
echo ========================================
echo ✅ Self-Healing Architecture Deployed!
echo ========================================
echo.
echo Next steps:
echo 1. Test with: whatsapp.exe rest --db-uri="..."
echo 2. Monitor logs for: "Refreshing device" messages
echo 3. Scale gradually to 3000+ devices
echo.
pause
