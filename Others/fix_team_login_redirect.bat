@echo off
echo ========================================
echo Fixing Team Login Redirect Issue
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
git commit -m "Fix team member login redirect issue - routes now properly bypass admin auth"

echo.
echo Pushing to GitHub...
git push origin main

echo.
echo ========================================
echo Fix deployed! Railway will auto-deploy.
echo ========================================
pause
