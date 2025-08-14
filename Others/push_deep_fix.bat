@echo off
echo Pushing deep template fix...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo [1/4] Adding files...
git add -A

echo [2/4] Committing...
git commit -m "Fix: Deep template fix - removed double input tags and malformed HTML"

echo [3/4] Pushing to main...
git push origin main

if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Push failed!
    pause
    exit /b 1
)

echo [4/4] Success!
echo.
echo Template rendering error fixed - removed double input tags.
echo.
pause
