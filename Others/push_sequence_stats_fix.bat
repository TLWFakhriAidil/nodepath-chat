@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo Committing changes...
git commit -m "Fix: Sequence Summary statistics calculation" -m "- Fixed status counts to use COUNT(DISTINCT recipient_phone) instead of COUNT" -m "- Ensure Total Should Send = Done + Failed + Remaining (not a separate count)" -m "- This fixes the issue where totals were not matching" -m "- Added debug logging to track statistics calculation"

echo Pushing to GitHub...
git push origin main

echo Done!
pause
