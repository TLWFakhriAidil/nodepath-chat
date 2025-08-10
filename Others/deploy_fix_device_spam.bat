@echo off
echo ========================================
echo Fixing Non-Existent Device Spam Issues
echo ========================================

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Building first to ensure no errors...
cd src
set CGO_ENABLED=0
go build -o ..\whatsapp.exe .
if %ERRORLEVEL% NEQ 0 (
    echo Build failed! Please fix errors before pushing.
    pause
    exit /b 1
)
cd ..
echo Build successful!

echo.
echo Adding all changes...
git add -A

echo Creating commit...
git commit -m "Fix: Stop spam for non-existent devices and enable new device processing

- Added DeviceCleanupManager to track cleaned devices and prevent spam
- Enhanced ensureWorker to check if device exists before creating worker
- Automatically clean up Redis queues for deleted devices
- Improved checkPendingQueues to validate devices and reduce logging
- Fixed lock acquisition issues
- System now properly handles deleted devices without spamming logs
- New devices (like 26bca561-8317-43c9-81ae-c820a5339513) will work properly"

echo Pushing to GitHub...
git push origin main

echo.
echo ========================================
echo Push Complete!
echo ========================================
echo.
echo Key improvements:
echo 1. No more spam for deleted device 3472b8c5-974b-4deb-bab9-792cc5a09c57
echo 2. Automatic Redis cleanup for non-existent devices
echo 3. New device 26bca561-8317-43c9-81ae-c820a5339513 will process campaigns
echo 4. Reduced logging spam - only logs important events
echo 5. Better device validation before worker creation
echo.
echo Railway will auto-deploy these changes.
echo.
pause
