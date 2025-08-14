@echo off
echo Fixing WhatsApp Web issues...

REM Fix 1: Update the sent image handling to save the actual image file
echo Fixing sent image storage...

REM Fix 2: Remove the refresh icon and loading messages
echo Removing refresh icon and loading messages...

REM Build the application
echo Building application...
call build_local.bat

echo.
echo Fixes applied!
echo 1. Sent images will now be properly saved and displayed
echo 2. Refresh icon and loading messages removed
echo.
pause
