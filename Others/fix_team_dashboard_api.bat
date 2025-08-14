@echo off
echo ========================================
echo Fixing Team Dashboard API Access
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
git commit -m "Fix team dashboard API access - add team endpoints and fix 401 errors"

echo.
echo Pushing to GitHub...
git push origin main

echo.
echo ========================================
echo Fix deployed! Railway will auto-deploy.
echo ========================================
pause
