@echo off
echo Pushing repository fixes for niche and schedule_time...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo Committing changes...
git commit -m "Fix repository to properly save and retrieve niche and schedule_time

- Updated GetSequences query to SELECT niche, status, schedule_time columns
- Fixed Scan to read all fields including niche and schedule_time
- Updated CreateSequence INSERT to include all fields
- Using correct column name 'schedule_time' not 'time_schedule'
- This fixes data population for niche and schedule_time display"

echo Pushing to main branch...
git push origin main

echo Push complete!
pause
