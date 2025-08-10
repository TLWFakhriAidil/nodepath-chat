@echo off
REM Run WhatsApp Stress Tests

echo ========================================
echo WhatsApp Multi-Device Stress Testing
echo ========================================
echo.

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\testing"

REM Build stress test
echo Building stress test...
go build -o stress_test.exe stress_test.go

if %ERRORLEVEL% NEQ 0 (
    echo ERROR: Build failed!
    pause
    exit /b 1
)

echo.
echo Starting stress tests...
echo.
stress_test.exe

pause
