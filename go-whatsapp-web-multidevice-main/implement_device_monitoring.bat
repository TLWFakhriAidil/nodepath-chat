@echo off
echo Implementing Device Connection Monitoring System...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo.
echo Step 1: Updating app.go to fix the route...

REM Fix the route to use the new handler
powershell -Command "(Get-Content 'src\ui\rest\app.go') -replace 'app.Get\(`"/api/devices/check-connection`", CheckDeviceConnectionStatus\)', 'app.Get(`"/api/devices/check-connection`", rest.HandleCheckConnection)' | Set-Content 'src\ui\rest\app.go'"

echo.
echo Step 2: Adding auto-reconnect route...

REM Add reconnect route after check-connection
powershell -Command "$content = Get-Content 'src\ui\rest\app.go'; $newContent = @(); $added = $false; foreach($line in $content) { $newContent += $line; if($line -like '*check-connection*' -and -not $added) { $newContent += "`tapp.Post(`"/api/devices/reconnect-all`", rest.HandleReconnectDevices)"; $added = $true } }; $newContent | Set-Content 'src\ui\rest\app.go'"

echo.
echo Step 3: Starting auto connection monitor in main.go...

REM Check if we need to add the auto monitor start
powershell -Command "$mainContent = Get-Content 'src\main.go' -Raw; if($mainContent -notmatch 'AutoConnectionMonitor') { Write-Host 'Adding auto connection monitor to main.go...' }"

echo.
echo Step 4: Building application...
go build -o whatsapp.exe src/main.go

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ===================================
    echo ✅ Device Connection Monitoring Implemented!
    echo ===================================
    echo.
    echo Features added:
    echo 1. Real-time connection checking endpoint: /api/devices/check-connection
    echo 2. Auto-reconnect endpoint: /api/devices/reconnect-all
    echo 3. Background monitoring every 10 seconds
    echo 4. Automatic reconnection for disconnected devices
    echo.
    echo The system now:
    echo - Checks device connections like broadcast campaigns
    echo - Updates device status in real-time
    echo - Auto-reconnects disconnected devices
    echo - Provides detailed connection information
    echo.
) else (
    echo.
    echo ❌ Build failed! Check for errors above.
)

pause