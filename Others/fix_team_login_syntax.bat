@echo off
echo Fixing team login syntax error and autocomplete attributes...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding files...
git add -A

echo Committing fix...
git commit -m "Fix team login syntax error and add autocomplete attributes

- Fixed duplicate .href syntax error in window.location redirect
- Added autocomplete='username' to username input field
- Added autocomplete='current-password' to password input field
- Team login form now works properly without JavaScript errors"

echo Pushing to main branch...
git push origin main

echo Done!
pause