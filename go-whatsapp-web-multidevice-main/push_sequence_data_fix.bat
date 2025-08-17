@echo off
echo Pushing fixes for sequence data and summary error...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo Committing changes...
git commit -m "Fix sequence data loading and summary error

- Fixed displaySequenceSummary null check to prevent TypeError
- Added proper null/undefined handling for summary data
- Fixed GetSequenceByID to SELECT all fields including niche and schedule_time
- This should fix both the empty sequences page and summary error"

echo Pushing to main branch...
git push origin main

echo Push complete!
pause
