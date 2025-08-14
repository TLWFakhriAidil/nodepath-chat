@echo off
echo Resetting to commit 315cf10 and pushing to main...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo [1/4] Resetting to commit 315cf10...
git reset --hard 315cf10

if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Reset failed!
    pause
    exit /b 1
)

echo [2/4] Current commit:
git log --oneline -1

echo [3/4] Force pushing to main branch...
git push origin main --force

if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Push failed!
    pause
    exit /b 1
)

echo [4/4] Push successful!
echo.
echo Repository has been reset to:
echo 315cf10 - Fix: Sequence steps with images - corrected column order and caption handling
echo.
echo This removes the Railway upload changes and keeps the image handling fixes.
echo.
pause
