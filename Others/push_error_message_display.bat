@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo Committing changes...
git commit -m "Add Error Message column to Failed Leads popup" -m "- Added 'Error Message' column header to lead details modal" -m "- Updated displayLeadDetails to show error_message from API response" -m "- Updated all colspan values from 4 to 5 to match new column count" -m "- Now users can see why messages failed (e.g., 'number not on WhatsApp')"

echo Pushing to GitHub...
git push origin main

echo Done!
pause
