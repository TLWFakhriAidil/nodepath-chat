@echo off
echo Pushing fix to match working version query...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo Committing changes...
git commit -m "Fix sequences query to match working version 68fec38

- Simplified GetSequences query to only select columns that exist
- Added COALESCE for schedule_time to handle NULL values
- Added detailed error logging to debug issues
- Set default status if not in database
- This should fix the no data issue"

echo Pushing to main branch...
git push origin main

echo Push complete!
pause
