@echo off
echo Pushing WhatsApp Campaign Database Fixes to GitHub...
echo.

cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo.
echo Committing changes...
git commit -m "Fix database structure and campaign/sequence functionality

- Changed campaigns.scheduled_time from TIME to TIMESTAMP for proper date/time handling
- Added min_delay_seconds and max_delay_seconds to campaigns table
- Added schedule_time, min_delay_seconds, max_delay_seconds to sequences table
- Added schedule_time to sequence_steps table
- Updated models to handle timestamp properly with custom JSON marshalling
- Fixed campaign display on calendar with proper date parsing
- Added delay configuration fields to campaign creation/edit forms
- Updated repository and REST handlers to support new fields
- Fixed campaign trigger to work with timestamp
- Added proper default values for delay settings
- Updated domain types for sequences to include all missing fields"

echo.
echo Pushing to main branch...
git push origin main

echo.
echo Done! All changes have been pushed to GitHub.
pause
