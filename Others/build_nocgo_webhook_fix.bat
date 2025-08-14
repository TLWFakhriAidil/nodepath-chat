@echo off
echo ============================================
echo Building WhatsApp Multi-Device (No CGO)
echo Webhook Update: Prevent Duplicate Devices
echo ============================================
echo.

REM Apply the webhook fix first
call apply_webhook_fix.bat

echo.
echo Building application...
cd src

REM Set environment variables for Windows build without CGO
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64

REM Build the application
go build -o ..\whatsapp.exe main.go

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ========================================
    echo BUILD SUCCESSFUL!
    echo ========================================
    echo.
    echo Executable: whatsapp.exe
    echo.
    echo What's new:
    echo - Webhook checks device_name before creating new device
    echo - Updates JID if device with same name exists
    echo - Prevents duplicate devices
    echo.
    cd ..
) else (
    echo.
    echo ========================================
    echo BUILD FAILED!
    echo ========================================
    echo Please check the error messages above.
    cd ..
    pause
    exit /b 1
)

echo Ready to test the updated webhook!
echo.
echo Test with:
echo curl -X POST http://localhost:3000/webhook/lead/create \
echo   -H "Content-Type: application/json" \
echo   -d "{\"name\":\"Test User\",\"phone\":\"60123456789\",\"device_id\":\"ABC123\",\"user_id\":\"your-user-id\",\"device_name\":\"TestDevice1\"}"
echo.
pause