@echo off
echo Reverting proxy implementation...

cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo.
echo Creating revert commit...
git revert HEAD --no-edit

echo.
echo Pushing revert to GitHub...
git push origin main

echo.
echo ✅ Successfully reverted proxy implementation!
echo.
echo Ready to work on other features now.
pause