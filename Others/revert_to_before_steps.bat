@echo off
echo Reverting to commit 68fec38 (before sequence step fixes)...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Current status:
git status

echo.
echo Reverting to 68fec38...
git reset --hard 68fec38

echo.
echo Force pushing to main branch...
git push origin main --force

echo.
echo Revert complete!
echo Current commit:
git log --oneline -1

pause
