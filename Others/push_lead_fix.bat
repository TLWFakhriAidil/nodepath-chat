@echo off
echo Building and pushing lead isolation fixes...
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
git commit -m "CRITICAL FIX: Lead isolation by device and user

- Fixed GetLeadsByDevice to properly filter by device ID
- Added GetLeadsByUserNicheAndStatus to prevent cross-user data leakage
- Campaigns now only see leads belonging to the user
- Leads are now properly isolated by device
- Fixed security issue where all users could see all leads
- This fixes campaigns showing 0 messages and lead display issues"

git push origin main
echo Push complete!
echo.
echo CRITICAL FIXES:
echo 1. Leads now properly filtered by device ID
echo 2. Campaigns only process user's own leads
echo 3. No more cross-user data visibility
echo 4. Your new device will only show its own leads
pause
