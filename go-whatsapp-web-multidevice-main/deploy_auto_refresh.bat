@echo off
REM Build and push auto device refresh feature

echo ========================================
echo Building Auto Device Refresh Feature
echo ========================================
echo.

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src"

REM Build without CGO
echo Building application...
set CGO_ENABLED=0
go build -o ../whatsapp.exe .

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo ERROR: Build failed!
    pause
    exit /b 1
)

echo.
echo Build successful!
echo.

cd ..

REM Git operations
echo Adding changes to git...
git add -A

echo.
echo Committing changes...
git commit -m "Add auto device refresh feature - automatically triggers refresh when device connection not found"

echo.
echo Pushing to GitHub...
git push origin main

echo.
echo ========================================
echo Auto Device Refresh Feature Deployed!
echo ========================================
echo.
echo When the error occurs:
echo "Failed to get device connection: no device connection found for device X"
echo.
echo The system will:
echo 1. Automatically trigger a device refresh
echo 2. Update device status to 'refreshing'
echo 3. Attempt to reconnect using stored JID
echo 4. Log the refresh attempt
echo.
pause
