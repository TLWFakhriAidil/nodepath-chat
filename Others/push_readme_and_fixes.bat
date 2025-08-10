@echo off
echo Updating README and pushing all fixes...
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
git add -A
git commit -m "Update README and fix message processing pipeline

- Updated README with latest fixes and improvements
- Fixed Redis to Worker queue bridge
- Messages now properly send via WhatsApp
- Device-specific lead isolation working
- Complete documentation of working flow"

git push origin main
echo Push complete!
pause
