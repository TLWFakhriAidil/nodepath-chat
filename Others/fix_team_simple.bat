@echo off
echo ========================================
echo Fixing Team Login Access and Device Matching
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
git commit -m "Fix team login redirect and use simple device name matching"

echo.
echo Pushing to GitHub...
git push origin main

echo.
echo ========================================
echo Fix deployed! Railway will auto-deploy.
echo.
echo Now team members will:
echo 1. Access /team-login without redirect
echo 2. See devices where device_name matches their username
echo 3. Have the same professional UI as admin
echo ========================================
pause
