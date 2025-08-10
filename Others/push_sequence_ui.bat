@echo off
echo Pushing Sequence Progress UI improvements...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo [1/4] Adding files...
git add -A

echo [2/4] Committing...
git commit -m "Fix: Sequence Progress UI - back button to sequences, beautiful gradient cards, broadcast message count, remove empty box"

echo [3/4] Pushing to main...
git push origin main

if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Push failed!
    pause
    exit /b 1
)

echo [4/4] Success!
echo.
echo Sequence Progress improvements:
echo - Back button returns to Sequences (not home)
echo - Beautiful gradient cards with hover effects
echo - Shows actual broadcast messages sent
echo - Cleaner UI without empty boxes
echo.
pause
