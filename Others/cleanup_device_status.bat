@echo off
echo Cleaning up device status implementation...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo.
echo Step 1: Removing auto connection monitor...
del src\infrastructure\whatsapp\auto_connection_monitor.go 2>nul

echo.
echo Step 2: Simplifying check-connection endpoint...

REM Create simple endpoint that just returns device status
echo package rest > src\ui\rest\check_connection_simple.go
echo. >> src\ui\rest\check_connection_simple.go
echo import ( >> src\ui\rest\check_connection_simple.go
echo     "github.com/gofiber/fiber/v2" >> src\ui\rest\check_connection_simple.go
echo     "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils" >> src\ui\rest\check_connection_simple.go
echo ) >> src\ui\rest\check_connection_simple.go
echo. >> src\ui\rest\check_connection_simple.go
echo // HandleCheckConnection returns simple success for frontend compatibility >> src\ui\rest\check_connection_simple.go
echo func HandleCheckConnection(c *fiber.Ctx) error { >> src\ui\rest\check_connection_simple.go
echo     return c.JSON(utils.ResponseData{ >> src\ui\rest\check_connection_simple.go
echo         Status:  200, >> src\ui\rest\check_connection_simple.go
echo         Code:    "SUCCESS", >> src\ui\rest\check_connection_simple.go
echo         Message: "Check connection endpoint", >> src\ui\rest\check_connection_simple.go
echo     }) >> src\ui\rest\check_connection_simple.go
echo } >> src\ui\rest\check_connection_simple.go

echo.
echo Step 3: Fixing device status updates to use only online/offline...

REM Update device handler to use simple status
powershell -Command "(Get-Content 'src\infrastructure\whatsapp\device_handler.go') -replace 'UpdateDeviceStatus\(deviceID, `"connected`"', 'UpdateDeviceStatus(deviceID, `"online`"' | Set-Content 'src\infrastructure\whatsapp\device_handler.go'"

REM Update all status checks to binary
powershell -Command "(Get-Content 'src\usecase\optimized_campaign_trigger.go') -replace 'device\.Status == `"connected`" \|\| device\.Status == `"Connected`" \|\|[\s]*device\.Status == `"online`" \|\| device\.Status == `"Online`"', 'device.Status == `"online`"' | Set-Content 'src\usecase\optimized_campaign_trigger.go'"

echo.
echo Step 4: Update broadcast processor...
powershell -Command "(Get-Content 'src\usecase\optimized_broadcast_processor.go') -replace 'device\.Status != `"online`" && device\.Status != `"Online`" &&[\s]*device\.Status != `"connected`" && device\.Status != `"Connected`"', 'device.Status != `"online`"' | Set-Content 'src\usecase\optimized_broadcast_processor.go'"

echo.
echo Step 5: Building application...
go build -o whatsapp.exe src/main.go

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ========================================
    echo ✅ Cleanup Complete!
    echo ========================================
    echo.
    echo What was removed:
    echo - Auto connection monitor (not needed)
    echo - Complex status values
    echo - Redundant checking
    echo.
    echo What remains:
    echo - Simple online/offline status
    echo - Existing campaign/broadcast/sequence checks
    echo - Real-time event updates
    echo.
    echo The system now uses existing checks at:
    echo - Campaign processing time
    echo - Broadcast every 2 seconds
    echo - Sequence every 15 seconds
    echo - QR/disconnect events
    echo.
) else (
    echo.
    echo ❌ Build failed! Check errors above.
)

pause