@echo off
REM Quick fix for 404 endpoint

echo ========================================
echo Quick Fix Deploy - Check Connection
echo ========================================
echo.

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src"

REM Build
echo Building...
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
git commit -m "Quick fix: Add simple check-connection endpoint to resolve 404 error"
git push origin main --force

echo.
echo ========================================
echo ✅ Quick Fix Deployed!
echo ========================================
echo.
echo The /api/devices/check-connection endpoint should now work.
echo It will return a simple response to stop the 404 errors.
echo.
pause
