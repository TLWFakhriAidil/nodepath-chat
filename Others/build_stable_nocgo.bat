@echo off
echo ========================================
echo Building WhatsApp Multi-Device (NO CGO)
echo ========================================
echo.

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src"

REM Set environment variables for no CGO
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64

echo Building with stability improvements...
go build -ldflags="-s -w" -o ..\whatsapp.exe .

if %ERRORLEVEL% NEQ 0 (
    echo Build failed! Please check the error messages above.
    cd /d ..
    pause
    exit /b 1
)

cd /d ..

echo.
echo ========================================
echo Build successful!
echo ========================================
echo.
echo Stability improvements added:
echo - Smart keepalive for null platform devices only
echo - 45-90 second random keepalive intervals
echo - Activity tracking to prevent keepalive during messaging
echo - 3-minute grace period before marking devices offline
echo - Auto-reconnect always enabled
echo - Extra pre-keys for connection stability
echo.
echo Run with: whatsapp.exe rest --db-uri="postgresql://..."
echo.
