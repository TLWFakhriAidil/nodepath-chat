@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
git add src/cmd/rest.go
git commit -m "Increase auto-reconnect delay to 60 seconds"
git push origin main
echo Done!
pause