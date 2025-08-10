@echo off
echo Fixing WhatsApp Client Registration Issue...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo.
echo Step 1: Adding improved diagnostics to diagnose endpoint...

REM Create a backup of app.go
copy "src\ui\rest\app.go" "src\ui\rest\app.go.backup" >nul 2>&1

echo.
echo Step 2: Building the project (without CGO)...
cd src
set CGO_ENABLED=0
go mod tidy
go build -o ../whatsapp.exe main.go

if %ERRORLEVEL% neq 0 (
    echo.
    echo Build failed! Please check the error messages above.
    pause
    exit /b 1
)

echo.
echo Step 3: Committing changes...
cd ..
git add -A
git commit -m "Fix WhatsApp client registration issue

- Added DiagnoseClients() for better debugging
- Added TryRegisterDeviceFromDatabase() to auto-register devices
- Fixed connection session tracking to use GetAllConnectionSessions()
- Added SetGlobalClient() to track the global WhatsApp instance
- Enhanced diagnose endpoint with client manager diagnostics
- Added auto-registration attempt for online devices without clients"

echo.
echo Step 4: Pushing to main branch...
git push origin main --force

echo.
echo ✅ Fix complete! The changes have been pushed.
echo.
echo What was fixed:
echo - WhatsApp client was not being registered in ClientManager
echo - Connection session tracking was not working properly
echo - Added diagnostics to better understand client registration
echo - Added auto-recovery for online devices without clients
echo.
echo Next steps:
echo 1. Wait for Railway to auto-deploy (1-2 minutes)
echo 2. Run the diagnostics endpoint again
echo 3. The client_manager section will show all registered clients
echo 4. If device is online, it should auto-register the client
echo.
pause
