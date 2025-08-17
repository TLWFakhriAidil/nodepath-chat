@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
git add -A
git commit -m "Trigger Railway auto-deploy - Device filter fix included"
git push origin main
echo.
echo ============================================
echo Push completed to trigger Railway deployment!
echo ============================================
echo.
echo Check Railway dashboard for deployment status.
echo The fix for device filter IS in this code.
echo.
pause