@echo off
echo ====================================
echo Building ULTRA STABLE WhatsApp Multi-Device
echo NO RATE LIMITS - MAXIMUM SPEED
echo ====================================

REM Set ultra stable environment variables
set ULTRA_STABLE_MODE=true
set IGNORE_RATE_LIMITS=true
set MAX_SPEED_MODE=true
set DISABLE_DELAYS=true
set FORCE_ONLINE_STATUS=true
set FORCE_RECONNECT_ATTEMPTS=100
set KEEP_ALIVE_INTERVAL=5

REM Clean old build
echo Cleaning old build...
if exist whatsapp.exe del whatsapp.exe

REM Build with no CGO for compatibility
echo Building with ULTRA STABLE features...
cd src
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -o ../whatsapp.exe -ldflags="-s -w" .
cd ..

if exist whatsapp.exe (
    echo ====================================
    echo BUILD SUCCESSFUL!
    echo ====================================
    echo ULTRA STABLE features enabled:
    echo - Devices will NEVER disconnect
    echo - NO rate limits applied
    echo - MAXIMUM speed messaging
    echo - ALL delays disabled
    echo - Devices forced online
    echo ====================================
    echo Run: whatsapp.exe
) else (
    echo BUILD FAILED!
    echo Check the error messages above.
)
