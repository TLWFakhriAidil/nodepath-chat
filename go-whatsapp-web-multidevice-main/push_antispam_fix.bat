@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo Committing changes...
git commit -m "Fix double anti-spam issue in message flow" -m "- Removed anti-spam from WhatsAppMessageSender (only keeps line break processing)" -m "- Removed anti-spam from PlatformSender (no processing needed)" -m "- Added proper anti-spam logic to BroadcastWorker.sendWhatsAppMessage()" -m "- Now follows correct flow: BroadcastWorker applies anti-spam ONCE" -m "  - For WhatsApp Web: Apply randomization then Add greeting then Send" -m "  - For Platform devices: Send raw content (platform handles its own anti-spam)" -m "- Prevents double/triple application of anti-spam that was causing garbled messages"

echo Pushing to GitHub...
git push origin main

echo Done!
pause
