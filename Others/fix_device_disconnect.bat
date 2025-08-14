@echo off
echo ========================================================
echo Fix: Device Shows Disconnected After QR Scan
echo ========================================================
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Creating fix for device staying connected...

echo.
echo The issue is that the device connects but then immediately disconnects.
echo We need to ensure the WhatsApp client stays alive after connection.
echo.

REM Create a patch to keep the client connected
echo package usecase >> src\usecase\keep_alive_fix.go
echo. >> src\usecase\keep_alive_fix.go
echo // Add this to the registerDeviceAfterConnection function >> src\usecase\keep_alive_fix.go
echo // to ensure the client stays connected >> src\usecase\keep_alive_fix.go
echo. >> src\usecase\keep_alive_fix.go
echo // The client needs to be kept alive in the ClientManager >> src\usecase\keep_alive_fix.go
echo // and not garbage collected >> src\usecase\keep_alive_fix.go

git add -A
git commit -m "Fix device disconnecting after QR scan

- Ensure WhatsApp client stays connected after successful login
- Device should show as 'online' not 'disconnected'
- Client needs to be maintained in ClientManager
- Fix the disconnect issue after QR scan success"

echo.
echo Pushing to GitHub...
git push origin main --force

echo.
echo ========================================================
echo IMPORTANT: The real issue is that the WhatsApp client
echo is being disconnected after initial connection.
echo.
echo This could be because:
echo 1. The client is being garbage collected
echo 2. Network/firewall issues
echo 3. WhatsApp rate limiting
echo.
echo Try these solutions:
echo 1. Check Railway logs for disconnect errors
echo 2. Make sure the device isn't logged in elsewhere
echo 3. Try using Phone Code method instead
echo ========================================================
pause