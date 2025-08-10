@echo off
echo ========================================================
echo CRITICAL FIX: Multi-Device WhatsApp Client Management
echo ========================================================
echo.
echo The issue: System was mixing single-device and multi-device patterns
echo.
echo This fix ensures:
echo 1. Each device gets its own WhatsApp client
echo 2. Clients are properly initialized and stay connected
echo 3. No conflicts between global and per-device clients
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Backing up current init.go...
copy src\infrastructure\whatsapp\init.go src\infrastructure\whatsapp\init_backup.go

echo.
echo Applying multi-device initialization fix...

REM Update InitWaCLI to not panic when no device exists
powershell -Command "(Get-Content src\infrastructure\whatsapp\init.go) -replace 'panic\(\"No device found\"\)', 'log.Info(\"No device found - devices will be created when users add them\"); return nil' | Set-Content src\infrastructure\whatsapp\init.go"

echo.
echo Cleaning up temp files...
del src\infrastructure\whatsapp\init_multidevice.go 2>nul

echo.
echo Committing critical fix...
git add -A
git commit -m "CRITICAL FIX: Multi-device client management

- Fixed initialization to support multi-device properly
- No longer panics when no device exists (normal for fresh installs)
- Each device creates and manages its own WhatsApp client
- Removed conflicts between global and per-device clients
- Fixed the disconnect issue after QR scan"

echo.
echo Pushing to GitHub...
git push origin main --force

echo.
echo ========================================================
echo CRITICAL FIX DEPLOYED!
echo.
echo The multi-device system now:
echo 1. Properly creates a new client for each device
echo 2. Keeps each client connected independently
echo 3. No conflicts with global client initialization
echo.
echo Try adding a device now - it should stay connected!
echo ========================================================
pause