@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Force pushing to main branch...
git push --force origin main

echo.
echo Creating/updating master branch...
git checkout -B master
git push --force origin master

echo.
echo Switching back to main branch...
git checkout main

echo.
echo Done! Both main and master branches have been force pushed.
