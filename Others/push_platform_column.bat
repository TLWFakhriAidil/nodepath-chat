@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo Committing changes...
git commit -m "Add Platform column to device-wise reports" -m "Frontend changes:" -m "- Added Platform column to campaign device report table" -m "- Added Platform column to sequence device report table" -m "- Shows platform badge (e.g., 'Wablas', 'Whacenter', or 'WhatsApp Web')" -m "" -m "Backend changes:" -m "- Updated queries to fetch platform field from user_devices table" -m "- Added platform field to DeviceReport and DeviceStepReport structs" -m "- Platform data now included in API responses for both campaign and sequence reports"

echo Pushing to GitHub...
git push origin main

echo Done!
pause
