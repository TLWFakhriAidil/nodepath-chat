@echo off
echo Pushing HTML template fix...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo [1/4] Adding files...
git add -A

echo [2/4] Committing...
git commit -m "Fix: HTML template error - fixed malformed input tags and image URL inputs"

echo [3/4] Pushing to main...
git push origin main

if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Push failed!
    pause
    exit /b 1
)

echo [4/4] Success!
echo.
echo Fixed template rendering error.
echo.
pause
