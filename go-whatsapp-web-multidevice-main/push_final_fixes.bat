@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
git add -A
git commit -m "Fix schedule_time display and status update in toggle

- Added time_schedule to field name checks for schedule_time display
- Fixed UpdateSequence to actually update the status field in database
- This fixes both schedule_time display and toggle status update"
git push origin main
pause
