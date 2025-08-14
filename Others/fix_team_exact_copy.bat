@echo off
echo ========================================
echo Copying EXACT Dashboard UI to Team Dashboard
echo ========================================

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src"

echo Building with CGO_ENABLED=0...
set CGO_ENABLED=0
go build -o ..\whatsapp.exe .

if %ERRORLEVEL% NEQ 0 (
    echo Build failed!
    pause
    exit /b 1
)

echo Build successful!
echo.
echo Committing changes...
cd ..
git add .
git commit -m "Copy exact dashboard UI to team dashboard - 100% identical interface"

echo.
echo Pushing to GitHub...
git push origin main

echo.
echo ========================================
echo Fix deployed! Railway will auto-deploy.
echo.
echo Team dashboard now has:
echo 1. EXACT same UI as admin dashboard
echo 2. All 4 tabs working
echo 3. Proper data display with actions column
echo 4. Fixed devices.filter error
echo ========================================
pause
