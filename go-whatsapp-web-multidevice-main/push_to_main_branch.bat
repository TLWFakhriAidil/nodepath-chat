@echo off
echo Pushing to main branch...

cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo.
echo Current branch:
git branch

echo.
echo Pushing to origin/main...
git push origin main

echo.
echo If that didn't work, let's try pushing current branch to main:
git push origin HEAD:main

echo.
echo Done! Check GitHub for the changes.
echo Repository: https://github.com/aqilrvsb/go-whatsapp-web-multidevice
pause