@echo off
echo Pushing sequence UI and functionality improvements...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo Committing changes...
git commit -m "Improve sequence UI and functionality

- Changed 'Tag' to 'Niche' in dashboard displays
- Fixed schedule_time field to use 'schedule_time' consistently
- New sequences now default to 'inactive' status
- Updated toggle function to switch between active/inactive only
- Removed paused/draft tabs, now only active/inactive
- Updated status colors: green for active, red for inactive
- Toggle button text changes to Activate/Deactivate
- Fixed backend to properly handle schedule_time field"

echo Pushing to main branch...
git push origin main

echo Push complete!
pause
