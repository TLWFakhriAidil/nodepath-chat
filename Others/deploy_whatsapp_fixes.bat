@echo off
echo 🚀 Deploying WhatsApp Web fixes...

cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

git add -A
git commit -m "🐛 Fix: WhatsApp Web button and authentication

- Moved WhatsApp Web from dropdown to main button (green button with WhatsApp icon)
- Fixed authentication issue - now properly checks session cookies
- No more redirect to login page when clicking WhatsApp Web
- Each connected device shows WhatsApp Web button prominently
- Updated README with latest changes"

git push origin main --force

echo ✅ WhatsApp Web fixes deployed!
pause
