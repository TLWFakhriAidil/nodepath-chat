@echo off
echo Committing and pushing URL-only image handling to GitHub...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo [1/4] Adding files to git...
git add -A

echo [2/4] Committing changes...
git commit -m "Fix: Convert image uploads to URL inputs only - no file uploads, just paste image URLs"

echo [3/4] Pushing to main branch...
git push origin main

if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Push failed!
    pause
    exit /b 1
)

echo [4/4] Push successful!
echo.
echo Changes pushed to GitHub main branch.
echo Railway will auto-deploy from main branch.
echo.
echo Image handling now works as follows:
echo - No file uploads - just URL inputs
echo - Paste image URLs (e.g., http://growrvsb.com/public/images/image.jpg)
echo - URLs stored directly in database
echo - No base64 encoding
echo - Works with sequences, campaigns, and AI campaigns
echo.
pause
