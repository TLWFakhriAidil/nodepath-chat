@echo off
echo ========================================
echo Pushing Device Logout Update to GitHub
echo ========================================
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Checking git status...
git status

echo.
echo Adding changes...
git add src/views/dashboard.html
git add device_logout_update.html
git add new_logout_function.js
git add device_management_complete.bat
git add DEVICE_LOGOUT_UPDATE_SUMMARY.md

echo.
echo Creating commit...
git commit -m "Remove Reset WhatsApp Session tab and enhance logout functionality

- Removed 'Reset WhatsApp Session' option from device dropdown menu
- Enhanced logout function to also remove WhatsApp session completely
- Users now only see 'Logout' which handles both disconnect and session removal
- Improved UX with SweetAlert2 dialogs instead of browser confirms
- Removed redundant resetDevice() function

This simplifies the device management UI and makes logout behavior more intuitive."

echo.
echo Pushing to main branch...
git push origin main

echo.
echo ========================================
echo Push completed!
echo ========================================
pause