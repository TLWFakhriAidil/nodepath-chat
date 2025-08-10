@echo off
echo ========================================
echo Pushing critical nil pointer fix
echo ========================================

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

REM Add all changed files
echo Adding changes...
git add src/infrastructure/whatsapp/multidevice/manager.go
git add src/infrastructure/whatsapp/multidevice/manager_fix.go
git add src/infrastructure/whatsapp/device_manager_init.go
git add src/infrastructure/whatsapp/multidevice_auto_reconnect.go

REM Commit with descriptive message
echo Committing changes...
git commit -m "Critical Fix: Nil pointer in DeviceManager store container

- Added ensureStoreContainer() method to check and reinitialize if nil
- Created manager_fix.go with retry logic for database connection
- Enhanced error handling and logging for debugging
- Added InitializeDeviceManagerWithRetry() with verification
- Store container now properly initialized before use
- Should fix the panic on Railway after restart"

REM Push to main branch
echo Pushing to main branch...
git push origin main

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✅ Successfully pushed critical fix!
    echo.
    echo This should resolve the nil pointer issue on Railway.
) else (
    echo.
    echo ❌ Push failed!
)

pause
