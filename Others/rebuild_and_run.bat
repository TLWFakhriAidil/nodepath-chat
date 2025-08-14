@echo off
echo ========================================
echo Stopping any running WhatsApp instances
echo ========================================

taskkill /F /IM whatsapp.exe 2>nul
taskkill /F /IM go-whatsapp-web-multidevice.exe 2>nul

echo.
echo ========================================
echo Pulling latest changes from GitHub
echo ========================================

git pull origin main

echo.
echo ========================================
echo Building WhatsApp Multi-Device (Local)
echo ========================================

set CGO_ENABLED=0
go build -o whatsapp.exe .

if %errorlevel% neq 0 (
    echo Build failed!
    pause
    exit /b 1
)

echo.
echo ========================================
echo Starting application with debug logging
echo ========================================

set DB_URI=postgresql://postgres:password@localhost:5432/whatsapp
set LOG_LEVEL=debug

echo Starting server...
whatsapp.exe rest --debug=true

pause
