@echo off
cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src"
set CGO_ENABLED=0
go build -o ..\whatsapp_modal_fix.exe 2>&1 | findstr /N "3480 3605"
pause
