@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo Committing changes...
git commit -m "Major Fix: Sequence Summary with Date Filters + Status Fixes" -m "Frontend Changes:" -m "- Added date range filter to Sequence Summary (like Campaign Summary)" -m "- Auto-filter to today's date on page load" -m "- Pass date filters to device reports" -m "- Change 'Campaign Device Report' to 'Sequence Device Report' for sequences" -m "- Change 'Campaign Details' to 'Sequence Details' for sequences" -m "- Add error_message column to failed leads popup" -m "" -m "Backend Changes:" -m "- Fixed ALL status queries from 'success' to 'sent'" -m "- Added date filter support to GetSequenceSummary API" -m "- Added date filter support to GetSequenceDeviceReport API" -m "- Fixed statistics calculation (done=sent with no error, failed, remaining=pending+queued)" -m "- Add error_message to lead details response" -m "" -m "Also updated README with all recent fixes"

echo Pushing to GitHub...
git push origin main

echo Done!
pause
