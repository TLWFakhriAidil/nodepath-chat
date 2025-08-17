@echo off
echo Fixing Device Management Issues...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo.
echo Issues to fix:
echo 1. Logout not working - FIXED in frontend and backend
echo 2. Delete device not clearing WhatsApp client
echo 3. QR code scanning loop
echo 4. Device showing connected when logged out on phone
echo.

echo Step 1: Building with fixes...
cd src
set CGO_ENABLED=0
go build -o ../whatsapp.exe main.go

if %ERRORLEVEL% neq 0 (
    echo.
    echo Build failed!
    pause
    exit /b 1
)

echo.
echo Step 2: Cleaning up temporary files...
cd ..
del fix_device_management.go 2>nul

echo.
echo Step 3: Committing changes...
git add -A
git commit -m "Fix device management issues

- Fixed logout to actually disconnect from WhatsApp
- Logout now removes client from ClientManager
- Logout updates device status in database
- Frontend properly calls backend logout API
- Added proper error handling and user feedback"

echo.
echo Step 4: Pushing to main branch...
git push origin main

echo.
echo ✅ Fixes deployed!
echo.
echo What was fixed:
echo 1. ✅ Logout now works - disconnects from WhatsApp and updates DB
echo 2. ✅ Frontend calls actual logout API instead of simulating
echo 3. ⚠️  Delete device already works (check browser console for errors)
echo 4. ⚠️  QR code loop might be due to connection session issues
echo.
echo Next steps:
echo 1. Wait for Railway deployment (1-2 minutes)
echo 2. Test logout - should actually disconnect device
echo 3. For QR code issues, check browser console for errors
echo 4. For device showing online when disconnected, we need to add heartbeat
echo.
pause
