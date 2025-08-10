@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo Committing changes...
git commit -m "Fix: Sequence Device Report title and statistics calculation" -m "Frontend changes:" -m "- Changed 'Campaign Device Report' to 'Sequence Device Report' for sequences" -m "- Changed 'Campaign Details' to 'Sequence Details' for sequences" -m "- Dynamically update modal titles based on context" -m "" -m "Backend changes:" -m "- Fixed status calculation to use 'sent' instead of 'success'" -m "- Added check for error_message IS NULL for done_send count" -m "- Changed 'pending' to include both 'pending' and 'queued' statuses" -m "- Fixed variable names from success/failed/pending to done_send/failed_send/remaining_send"

echo Pushing to GitHub...
git push origin main

echo Done!
pause
