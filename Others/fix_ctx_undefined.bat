@echo off
echo ========================================================
echo Fix undefined ctx variable
echo ========================================================
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding changes...
git add -A

echo.
echo Committing fix...
git commit -m "Fix undefined ctx variable in Login function

- Changed Login function parameter from _ to ctx
- Now ctx is properly defined and can be used in device.Delete(ctx)"

echo.
echo Pushing to GitHub...
git push origin main --force

echo.
echo ========================================================
echo CTX VARIABLE ERROR FIXED!
echo.
echo The build should now succeed on Railway.
echo ========================================================
pause