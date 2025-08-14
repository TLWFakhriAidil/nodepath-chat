@echo off
echo Reverting back to commit 68fec38...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Current status:
git status

echo.
echo Reverting to 68fec38 - Fix sequence creation: niche, time_schedule, and steps saving
git reset --hard 68fec38

echo.
echo Force pushing to main branch...
git push origin main --force

echo.
echo Revert complete!
echo Current commit:
git log --oneline -1

pause
