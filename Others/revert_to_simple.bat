@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
echo Reverting to simple campaign fix...

echo.
echo Force pushing to main branch...
git push --force origin main

echo.
echo Reverted successfully!
echo.
echo WHAT THIS VERSION HAS:
echo - Fixed SQL to use campaign_date instead of scheduled_date
echo - Added COALESCE for device_id to handle NULL
echo - Campaign display improvements (clickable names, delete icons)
echo - Day labels on calendar
echo.
echo WHAT YOU NEED TO DO:
echo 1. Run this SQL on your Railway database:
echo    UPDATE campaigns SET scheduled_time = NULL WHERE scheduled_time = 'Invalid Date';
echo.
echo 2. Login credentials remain:
echo    Email: admin@whatsapp.com
echo    Password: changeme123
echo.
echo That's it! No complex migrations needed.
echo.
pause
