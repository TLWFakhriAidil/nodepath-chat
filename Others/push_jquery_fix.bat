@echo off
echo Fixing jQuery references in AI Campaign modals...
cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
git add -A
git commit -m "Fix jQuery references - use vanilla JavaScript for Bootstrap 5 modals"
git push origin main
echo Push completed!