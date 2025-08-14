@echo off
echo Reverting to commit 643669b (niche fix)...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Fetching latest from remote...
git fetch origin

echo.
echo Reverting to 643669b - Fix sequence repository to include niche and all fields in queries
git reset --hard 643669b

echo.
echo Force pushing to main branch...
git push origin main --force

echo.
echo Revert complete!
echo Current commit:
git log --oneline -1

pause
