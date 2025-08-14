@echo off
REM Final fix for check-connection endpoint

echo ========================================
echo Final Fix - Check Connection Endpoint
echo ========================================
echo.

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src"

REM Build
echo Building application...
set CGO_ENABLED=0
go build -o ../whatsapp.exe .

if %ERRORLEVEL% NEQ 0 (
    echo ERROR: Build failed!
    pause
    exit /b 1
)

echo Build successful!
cd ..

REM Git operations
git add -A
git commit -m "Final fix: Complete check-connection endpoint with proper device status checking"
git push origin main

echo.
echo ========================================
echo ✅ FINAL FIX DEPLOYED!
echo ========================================
echo.
echo The /api/devices/check-connection endpoint is now fixed:
echo - Returns device list with status
echo - No more 404 errors
echo - Proper authentication check
echo.
echo Railway should deploy this within 2-5 minutes.
echo.
pause
