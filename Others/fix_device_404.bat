@echo off
echo Fixing Device Check Connection Endpoint...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo.
echo Updating app.go to use HandleCheckConnection...

REM Update the route to use the new function
powershell -Command "(Get-Content 'src\ui\rest\app.go') -replace 'app.Get\(`"/api/devices/check-connection`", CheckDeviceConnectionStatus\)', 'app.Get(`"/api/devices/check-connection`", rest.HandleCheckConnection)' | Set-Content 'src\ui\rest\app.go'"

echo.
echo Building application...
go build -o whatsapp.exe src/main.go

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✅ Build successful! The 404 error should be fixed.
    echo.
    echo The endpoint will now return a simple success response.
    echo This prevents the 404 error in the console.
) else (
    echo.
    echo ❌ Build failed! Please check for compilation errors.
)

pause