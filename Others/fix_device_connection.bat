@echo off
echo Fixing Device Connection Endpoint...

REM Fix the route registration
powershell -Command "(Get-Content 'src\ui\rest\app.go') -replace 'app.Get\(\"/api/devices/check-connection\", CheckDeviceConnectionStatus\)', 'app.Get(\"/api/devices/check-connection\", rest.CheckDeviceConnectionStatus)' | Set-Content 'src\ui\rest\app.go'"

echo Fix applied! Building...

REM Build the application
go build -o whatsapp.exe src/main.go

echo.
echo Fix complete! The endpoint should now work properly.
echo.
pause