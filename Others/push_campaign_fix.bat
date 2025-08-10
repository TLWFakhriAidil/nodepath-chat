@echo off
echo Committing and pushing campaign image input fixes...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo [1/4] Adding files to git...
git add -A

echo [2/4] Committing changes...
git commit -m "Fix: Campaign and AI Campaign image inputs - use URL inputs instead of file uploads"

echo [3/4] Pushing to main branch...
git push origin main

if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Push failed!
    pause
    exit /b 1
)

echo [4/4] Push successful!
echo.
echo All image uploads now converted to URL inputs:
echo - Sequences: URL input
echo - Campaigns: URL input  
echo - AI Campaigns: URL input
echo.
echo Just paste image URLs - no file uploads!
echo.
pause
