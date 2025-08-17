@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
git add -A
git commit -m "Fix foreign key constraint error: change campaign_id to INTEGER in broadcast_messages table"
git push origin main
echo.
echo Push completed!
pause
