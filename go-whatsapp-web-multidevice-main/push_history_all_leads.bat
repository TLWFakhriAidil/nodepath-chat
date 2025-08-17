@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo Committing changes...
git commit -m "Fix: History Triggers now shows ALL leads with historical triggers" -m "- Changed logic to find ALL phone numbers in broadcast_messages with triggers" -m "- Creates temporary lead entries for phones not in leads table" -m "- Shows complete history including leads that were never saved to leads table" -m "- Uses GROUP_CONCAT to get all triggers at once for better performance" -m "- Properly merges current and historical triggers"

echo Pushing to GitHub...
git push origin main

echo Done!
pause
