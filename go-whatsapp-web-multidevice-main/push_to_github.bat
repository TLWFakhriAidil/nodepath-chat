@echo off
echo ========================================
echo Pushing Sequence Fixes to GitHub
echo ========================================
echo.

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

echo Adding changes...
git add .

echo.
echo Committing changes...
git commit -m "Fix critical issues for both sequences and campaigns: duplicate prevention and message ordering

- Added duplicate checking for BOTH sequences and campaigns in QueueMessage()
  - Sequences: Check sequence_stepid, recipient_phone, and device_id
  - Campaigns: Check campaign_id, recipient_phone, and device_id
- Fixed message ordering to use scheduled_at instead of created_at
- Verified mutex locking in device workers prevents race conditions
- Updated README and documentation with complete fixes"

echo.
echo Pushing to main branch...
git push origin main

echo.
echo ========================================
echo Push complete!
echo ========================================
pause
