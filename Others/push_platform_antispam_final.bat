@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo Committing changes...
git commit -m "Fix: Apply anti-spam to ALL devices (including platform devices)" -m "- Platform devices now also get anti-spam applied (greeting + randomization)" -m "- Both WhatsApp Web and Platform devices use the same anti-spam logic" -m "- Removed the platform device check that was skipping anti-spam" -m "- This ensures consistent messaging across all device types"

echo Pushing to GitHub...
git push origin main

echo Done!
pause
