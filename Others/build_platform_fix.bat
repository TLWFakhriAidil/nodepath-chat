@echo off
cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src"
set CGO_ENABLED=0
go build -o ../whatsapp.exe .
