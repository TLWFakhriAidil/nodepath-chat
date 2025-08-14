@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
git add -A
git commit -m "Add alert to verify new version deployment"
git push origin main
echo.
echo Push completed! 
echo.
echo IMPORTANT: After Railway deploys, you should see an alert popup saying:
echo "NEW VERSION LOADED! v1.1.0 - If you see this, the deployment worked!"
echo.
echo If you don't see this alert, then Railway hasn't deployed the new version yet.
pause