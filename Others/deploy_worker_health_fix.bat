@echo off
echo ========================================
echo Deploying Worker Health Improvements
echo ========================================

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

:: Add all changes
git add -A

:: Commit with message
git commit -m "Fix: Worker health check and auto-reconnect system - All control buttons functional"

:: Push to main branch
git push origin main --force

echo.
echo ========================================
echo Deployment Complete!
echo ========================================
echo.
echo Summary of improvements:
echo - Device Health Monitor with auto-reconnect
echo - Enhanced Client Manager
echo - Improved Worker Health Checks
echo - All Worker Control buttons working
echo - Frontend integration complete
echo.
echo Railway will auto-deploy these changes!
echo.
pause
