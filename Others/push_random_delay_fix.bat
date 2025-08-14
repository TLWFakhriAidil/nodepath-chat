@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo Committing changes...
git commit -m "Fix: Implement true random delays for all message types" -m "- Fixed calculateRandomDelay to generate truly random delays between min and max" -m "- Previously always used middle value (e.g., 10-30 always gave 20)" -m "- Now properly randomizes delay for each message" -m "- Works for campaigns, sequences, platform devices, and WhatsApp Web" -m "- Helps avoid pattern detection and improves anti-spam"

echo Pushing to GitHub...
git push origin main

echo Done!
pause
