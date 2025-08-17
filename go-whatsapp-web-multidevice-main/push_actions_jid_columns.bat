@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo Committing changes...
git commit -m "Add Actions and JID columns to device reports" -m "Frontend changes:" -m "- Added Actions column with 'Resend Failed' button for sequence steps" -m "- Added JID column showing device JID with edit button" -m "- Both columns added to campaign and sequence device reports" -m "" -m "Backend changes:" -m "- Added ResendFailedSequenceStep endpoint to reset failed messages to pending" -m "- Added UpdateDeviceJID endpoint to update device JID" -m "- Added JID field to DeviceReport and DeviceStepReport structs" -m "- Updated queries to fetch JID from user_devices table" -m "" -m "Functionality:" -m "- Resend button finds all failed messages for that device/step and resets to pending" -m "- Edit JID button opens popup to update the device JID value" -m "- JID shown in truncated format with full value on hover"

echo Pushing to GitHub...
git push origin main

echo Done!
pause
