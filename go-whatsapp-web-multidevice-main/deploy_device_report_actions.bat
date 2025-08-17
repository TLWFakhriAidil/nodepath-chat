@echo off
echo Deploying Device Report Actions and Real-time Status Updates...
echo.

REM Build the application
echo Building application...
cd src
go build -o ../whatsapp.exe main.go
if errorlevel 1 (
    echo Build failed!
    pause
    exit /b 1
)
cd ..

REM Add and commit changes
echo.
echo Adding changes to git...
git add -A

echo.
echo Committing changes...
git commit -m "feat: Add device report actions and real-time status updates

- Added Actions column to device report with retry and transfer icons
- Transfer icon allows copying successful AI campaign leads to regular leads table
- Implemented real-time device connection status checking
- Added /api/devices/check-connection endpoint for status updates
- Fixed device status to show offline after logout
- Dashboard now checks real connection status before loading devices
- Ensures device status accurately reflects WhatsApp connection state"

REM Push to GitHub
echo.
echo Pushing to GitHub...
git push origin main

echo.
echo Deployment complete!
echo.
echo New features:
echo 1. Device Report Actions column with transfer capability
echo 2. Real-time device status updates
echo 3. Accurate connection status after logout
echo.
pause
