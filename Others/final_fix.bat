@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
git add -A
git commit -m "RESTORE dashboard.html with ALL FIXES - device filter, auth, proper spacing"
git push origin main
echo.
echo ============================================
echo CORRECT DASHBOARD RESTORED WITH ALL FIXES!
echo ============================================
echo.
echo Fixes included:
echo 1. Device filter: innerHTML = '<option value="all">All Devices</option>'
echo 2. Authentication: credentials: 'include' on all fetch calls
echo 3. Proper function spacing (no more undefined errors)
echo 4. Version logging for verification
echo.
echo Railway should now deploy the WORKING version!
echo.
pause