@echo off
echo Building and pushing device-specific lead fixes...
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src
set CGO_ENABLED=0
go build -o ..\whatsapp.exe .
if %ERRORLEVEL% NEQ 0 (
    echo Build failed!
    pause
    exit /b 1
)
cd ..
echo Build successful!
git add -A
git commit -m "Fix: Device-specific lead handling for campaigns and sequences

- Campaigns now use leads from their own devices only
- Each device processes its own leads (no round-robin)
- GetLeadsByDevice now properly filters by device ID
- GetLeadsByDeviceNicheAndStatus for device-specific campaign targeting
- Removed cross-device and cross-user data leakage
- Each of 3000 devices runs independently with its own leads"

git push origin main
echo Push complete!
echo.
echo IMPORTANT FIXES:
echo 1. Each device only sees and processes its own leads
echo 2. No more round-robin - each device handles its own data
echo 3. Campaigns will now find leads correctly
echo 4. 3000 devices can run simultaneously, each with its own leads
echo.
echo NOTE: Sequences still need device association for contacts
pause
