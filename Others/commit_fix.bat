@echo off
cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding changes...
git add -A

echo Committing...
git commit -m "Fix team dashboard niches endpoint URL"

echo Pushing...
git push origin main

echo Done!
pause
