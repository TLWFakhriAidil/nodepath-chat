@echo off
echo ========================================
echo Pushing Device Filter Fix
echo ========================================
echo.

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

REM Add changes
git add src/views/dashboard.html

REM Commit with proper message
git commit -m "Fix device filter to default to All Devices and group by name"

REM Push
git push origin main

echo.
echo ========================================
echo Push completed!
echo ========================================
pause
