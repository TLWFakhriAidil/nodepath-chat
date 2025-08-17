@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
git add -A
git commit -m "Fix GetSequences to actually SELECT and SCAN status column"
git push origin main
echo Done!
pause
